package tlsparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"github.com/stretchr/testify/assert"
)

func (mockService *mockService) buildConnectionsResponse(url string) ([]model.TLSConnection, error) {
	if strings.Contains(url, "scenario=error") {
		return nil, errors.New("error")
	}
	return repo, nil
}

func (mockService *mockService) buildDetailsResponse(url string) (model.TLSDetails, error) {
	if strings.Contains(url, "scenario=error") {
		return model.TLSDetails{}, errors.New("error")
	}
	return repoDetail, nil
}

func TestTLSParserConnectionsHandler(t *testing.T) {

	var tests = []struct {
		scenario string
		want     []model.TLSConnection
		status   int
		err      string
	}{
		{"ok", repo, http.StatusOK, ""},
		{"error", nil, http.StatusInternalServerError, "error"},
	}

	service := &mockService{}

	o11yController := O11yController{service}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/api/data?scenario=%s", test.scenario), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o11yController.TLSParserConnectionsHandler)
			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, rr.Code, test.status)

			var result []model.TLSConnection
			json.Unmarshal([]byte(rr.Body.String()), &result)

			assert.EqualValues(t, test.want, result)
		})
	}
}

func TestTLSParserConnectionDetailsHandler(t *testing.T) {
	var tests = []struct {
		scenario string
		id       string
		want     model.TLSDetails
		status   int
	}{
		{"ok", "id1", repoDetail, http.StatusOK},
		{"error", "id1", model.TLSDetails{}, http.StatusInternalServerError},
		{"empty_id", "", model.TLSDetails{}, http.StatusOK},
	}

	service := &mockService{}

	o11yController := O11yController{service}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/api/data/%s?scenario=%s", test.id, test.scenario), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o11yController.TLSParserConnectionDetailsHandler)
			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, rr.Code, test.status)

			var result model.TLSDetails
			json.Unmarshal([]byte(rr.Body.String()), &result)

			assert.EqualValues(t, test.want, result)
		})
	}
}
