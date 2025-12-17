package tlsparser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/stretchr/testify/assert"
)

var repo = []model.TLSConnection{
	{Id: "id1", Src: "src1", Domain: "k8spacket.io"},
	{Id: "id2", Src: "src2", Domain: "ebpf.io"},
}

var repoDetail = model.TLSDetails{Id: "id1", Domain: "k8spacket.io", UsedTLSVersion: "TLS 1.2"}

type mockTlsParserService struct {
	Service
	client, server     string
	domain, usedCipher string
	clientTLSVersions  []string
}

func (mockTlsParserService *mockTlsParserService) storeInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails) {

	mockTlsParserService.client = tlsConnection.Src
	mockTlsParserService.server = tlsConnection.Dst
	mockTlsParserService.domain = tlsConnection.Domain
	mockTlsParserService.usedCipher = tlsConnection.UsedCipherSuite
	mockTlsParserService.clientTLSVersions = tlsDetails.ClientTLSVersions
}

func (mockTlsParserService *mockTlsParserService) getConnection(id string) model.TLSDetails {
	if id == "not_found" {
		return model.TLSDetails{}
	}
	return repoDetail
}

func (mockTlsParserService *mockTlsParserService) filterConnections(query url.Values) []model.TLSConnection {
	return repo
}

func TestTLSConnectionHandler(t *testing.T) {

	service := &mockTlsParserService{}
	controller := &Controller{service: service}

	req, err := http.NewRequest("GET", "/tlsparser/connections/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.TLSConnectionHandler)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(t, rr.Code, http.StatusOK)

	var response []model.TLSConnection
	json.Unmarshal([]byte(rr.Body.String()), &response)

	assert.EqualValues(t, repo, response)

}

func TestTLSConnectionHandlerDetails(t *testing.T) {

	var tests = []struct {
		scenario string
		want     model.TLSDetails
		status   int
		error    string
	}{
		{"id1", repoDetail, http.StatusOK, ""},
		{"not_found", model.TLSDetails{}, http.StatusNotFound, "ala"},
	}

	service := &mockTlsParserService{}
	controller := &Controller{service: service}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/connections/%s", test.scenario), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TLSConnectionHandler)

			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, rr.Code, test.status)

			var response model.TLSDetails
			json.Unmarshal([]byte(rr.Body.String()), &response)

			assert.EqualValues(t, test.want, response)
		})
	}

}
