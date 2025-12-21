package o11y

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.TLSConnection{
	{Id: "id1", Src: "src1", Domain: "k8spacket.io"},
	{Id: "id2", Src: "src2", Domain: "ebpf.io"},
}

var dbDetails = model.TLSDetails{Id: "id1", UsedTLSVersion: "TLS 1.2", Domain: "k8spacket.io"}

type mockHttpClient struct {
	httpClient httpclient.Client
	scenario   string
}

func (httpClient *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if httpClient.scenario == "ok" {
		result, _ := json.Marshal(dbState)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if httpClient.scenario == "ok_detail" {
		result, _ := json.Marshal(dbDetails)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if httpClient.scenario == "ok_detail_empty" {
		result, _ := json.Marshal(model.TLSDetails{})
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if httpClient.scenario == "error" {
		var response []model.TLSConnection
		err := errors.New("error")
		result, _ := json.Marshal(response)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	if httpClient.scenario == "read" {
		reader := BrokenReader{}
		return &http.Response{
			Body:       &reader,
			StatusCode: http.StatusOK,
		}, nil
	}
	if httpClient.scenario == "parse" {
		result := []byte("parse error")
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	return &http.Response{}, nil
}

type BrokenReader struct{}

func (br *BrokenReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failed reading")
}

func (br *BrokenReader) Close() error {
	return fmt.Errorf("failed closing")
}

type mockK8SClient struct {
	k8sClient k8sclient.Client
}

func (k8sClient *mockK8SClient) GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string {
	return []string{"127.0.0.1"}
}

func TestTLSParserConnectionsHandler(t *testing.T) {

	var tests = []struct {
		scenario string
		want     []model.TLSConnection
		status   int
	}{
		{"ok", dbState, http.StatusOK},
		{"error", []model.TLSConnection{}, http.StatusOK},
		{"read", []model.TLSConnection{}, http.StatusOK},
		{"parse", []model.TLSConnection{}, http.StatusOK},
	}

	mockHttpClient := &mockHttpClient{}
	mockK8SClient := &mockK8SClient{}

	o11yController := NewO11yHandler(mockHttpClient, mockK8SClient)

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			mockHttpClient.scenario = test.scenario

			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/api/data?scenario=%s", test.scenario), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o11yController.TLSParserConnectionsHandler)
			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, test.status, rr.Code)

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
		{"ok_detail", "id1", dbDetails, http.StatusOK},
		{"ok_detail_empty", "id1", model.TLSDetails{}, http.StatusOK},
		{"error", "id1", model.TLSDetails{}, http.StatusOK},
		{"empty_id", "", model.TLSDetails{}, http.StatusOK},
		{"read", "", model.TLSDetails{}, http.StatusOK},
		{"parse", "", model.TLSDetails{}, http.StatusOK},
	}

	mockHttpClient := &mockHttpClient{}
	mockK8SClient := &mockK8SClient{}

	o11yController := NewO11yHandler(mockHttpClient, mockK8SClient)

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			mockHttpClient.scenario = test.scenario

			req, err := http.NewRequest("GET", fmt.Sprintf("/tlsparser/api/data/%s", test.id), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o11yController.TLSParserConnectionDetailsHandler)
			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, test.status, rr.Code)

			var result model.TLSDetails
			json.Unmarshal([]byte(rr.Body.String()), &result)

			assert.EqualValues(t, test.want, result)
		})
	}
}
