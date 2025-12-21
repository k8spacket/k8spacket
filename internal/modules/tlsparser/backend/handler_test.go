package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.TLSConnection{
	{Id: "id1", Src: "src1"},
	{Id: "id2", Src: "src2"},
}

var dbDetails = model.TLSDetails{Id: "id1", UsedTLSVersion: "TLS 1.2", UsedCipherSuite: "TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA"}

type mockRepository struct {
	repository.Repository
	resultConnection model.TLSConnection
	resultDetails    model.TLSDetails
	from, to         time.Time
	scenario         string
}

func (mockRepository *mockRepository) Query(from time.Time, to time.Time) []model.TLSConnection {
	mockRepository.from = from
	mockRepository.to = to
	return dbState
}

func (mockRepository *mockRepository) Read(key string) model.TLSDetails {
	if mockRepository.scenario == "not_found" {
		return model.TLSDetails{}
	}
	return dbDetails
}

func TestTLSConnectionHandler(t *testing.T) {

	mockRepository := &mockRepository{}
	handler := NewHandler(mockRepository)

	req, err := http.NewRequest("GET", "/tlsparser/connections/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(handler.TLSConnectionHandler)

	httpHandler.ServeHTTP(rr, req)

	assert.EqualValues(t, rr.Code, http.StatusOK)

	var response []model.TLSConnection
	json.Unmarshal([]byte(rr.Body.String()), &response)

	assert.EqualValues(t, dbState, response)

}

func TestTLSConnectionHandler_FilterConnections(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		scenario, from, to string
		wantFrom, wantTo   time.Time
		error              string
	}{
		{"correct", "1640998861000", "1675303322000", time.Time(time.Date(2022, time.January, 1, 1, 1, 1, 0, time.UTC)), time.Time(time.Date(2023, time.February, 2, 2, 2, 2, 0, time.UTC)), ""},
		{"wrong from", "wrong from", "1675303322000", time.Time{}, time.Time(time.Date(2023, time.February, 2, 2, 2, 2, 0, time.UTC)), "[api] cannot parse value"},
		{"wrong to", "1640998861000", "wrong to", time.Time(time.Date(2022, time.January, 1, 1, 1, 1, 0, time.UTC)), time.Time{}, "[api] cannot parse value"},
	}

	mockRepository := &mockRepository{}
	handler := NewHandler(mockRepository)

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			query := url.Values{}
			query.Add("from", test.from)
			query.Add("to", test.to)

			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/connections/?from=%s&to=%s", test.from, test.to), nil)
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println(req.URL.Query())

			rr := httptest.NewRecorder()
			httpHandler := http.HandlerFunc(handler.TLSConnectionHandler)

			httpHandler.ServeHTTP(rr, req)

			assert.EqualValues(t, rr.Code, http.StatusOK)

			assert.EqualValues(t, test.wantFrom, mockRepository.from)
			assert.EqualValues(t, test.wantTo, mockRepository.to)
			assert.Contains(t, str.String(), test.error)

		})
	}
}

func TestTLSConnectionHandlerDetails(t *testing.T) {

	var tests = []struct {
		scenario string
		want     model.TLSDetails
		status   int
	}{
		{"id1", dbDetails, http.StatusOK},
		{"not_found", model.TLSDetails{}, http.StatusNotFound},
	}

	mockRepository := &mockRepository{}
	handler := NewHandler(mockRepository)

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			mockRepository.scenario = test.scenario
			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/connections/%s", test.scenario), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			httpHandler := http.HandlerFunc(handler.TLSConnectionHandler)

			httpHandler.ServeHTTP(rr, req)

			assert.EqualValues(t, rr.Code, test.status)

			var response model.TLSDetails
			json.Unmarshal([]byte(rr.Body.String()), &response)

			assert.EqualValues(t, test.want, response)
		})
	}

}
