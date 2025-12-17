package tlsparser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/update"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.TLSConnection{
	{Id: "id1", Src: "src1"},
	{Id: "id2", Src: "src2"},
}

var dbDetails = model.TLSDetails{Id: "id1", UsedTLSVersion: "TLS 1.2"}

type mockRepository struct {
	repo             repository.Repository
	resultConnection model.TLSConnection
	resultDetails    model.TLSDetails
	from, to         time.Time
}

func (mockRepository *mockRepository) Query(from time.Time, to time.Time) []model.TLSConnection {
	mockRepository.from = from
	mockRepository.to = to
	return []model.TLSConnection{}
}

func (mockRepository *mockRepository) UpsertConnection(key string, value *model.TLSConnection) {
	mockRepository.resultConnection = *value
}

func (mockRepository *mockRepository) Read(key string) model.TLSDetails {
	return model.TLSDetails{UsedCipherSuite: "TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA"}
}

func (mockRepository *mockRepository) UpsertDetails(key string, value *model.TLSDetails, fn repository.Fn) {
	fn(value, &mockRepository.resultDetails)
	mockRepository.resultDetails = *value
}

type mockCertificateUpdater struct {
	update.Updater
	fnCalled bool
}

func (mockCertificateUpdater *mockCertificateUpdater) Update(newValue *model.TLSDetails, oldValue *model.TLSDetails) {
	mockCertificateUpdater.fnCalled = true
}

type mockHttpClient struct {
	httpClient httpclient.Client
}

func (httpClient *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Query().Get("scenario") == "ok" {
		result, _ := json.Marshal(dbState)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if req.URL.Query().Get("scenario") == "ok_detail" {
		result, _ := json.Marshal(dbDetails)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if req.URL.Query().Get("scenario") == "ok_detail_empty" {
		result, _ := json.Marshal(model.TLSDetails{})
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if req.URL.Query().Get("scenario") == "error" {
		response := []model.TLSConnection{}
		err := errors.New("error")
		result, _ := json.Marshal(response)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	if req.URL.Query().Get("scenario") == "read" {
		reader := BrokenReader{}
		return &http.Response{
			Body:       &reader,
			StatusCode: http.StatusOK,
		}, nil
	}
	if req.URL.Query().Get("scenario") == "parse" {
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

func TestStoreInDatabase(t *testing.T) {

	mockRepository := &mockRepository{}
	mockCertificateUpdater := &mockCertificateUpdater{}

	service := TlsParserService{repo: mockRepository, updater: mockCertificateUpdater, httpClient: &httpclient.HttpClient{}, k8sClient: &k8sclient.K8SClient{}}

	tlsConnection := model.TLSConnection{Src: "src"}
	tlsDetails := model.TLSDetails{UsedTLSVersion: "TLS 1.2"}

	service.storeInDatabase(&tlsConnection, &tlsDetails)

	assert.NotEmpty(t, mockRepository.resultConnection)
	assert.NotEmpty(t, mockRepository.resultDetails)

	assert.NotEmpty(t, mockRepository.resultConnection.Id)
	assert.NotEmpty(t, mockRepository.resultDetails.Id)

	assert.EqualValues(t, true, mockCertificateUpdater.fnCalled)

}

func TestRead(t *testing.T) {
	mockRepository := &mockRepository{}
	service := TlsParserService{repo: mockRepository, updater: &update.CertificateUpdater{}, httpClient: &httpclient.HttpClient{}, k8sClient: &k8sclient.K8SClient{}}

	result := service.getConnection("key")

	assert.EqualValues(t, "TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA", result.UsedCipherSuite)
}

func TestFilterConnections(t *testing.T) {

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

	service := TlsParserService{repo: mockRepository, updater: &update.CertificateUpdater{}, httpClient: &mockHttpClient{}, k8sClient: &k8sclient.K8SClient{}}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			query := url.Values{}
			query.Add("from", test.from)
			query.Add("to", test.to)

			service.filterConnections(query)

			assert.EqualValues(t, test.wantFrom, mockRepository.from)
			assert.EqualValues(t, test.wantTo, mockRepository.to)
			assert.Contains(t, str.String(), test.error)

		})
	}
}

func TestBuildConnectionsResponse(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		scenario string
		want     []model.TLSConnection
		error    string
	}{
		{"ok", dbState, ""},
		{"error", []model.TLSConnection{}, "[api] Cannot get stats"},
		{"read", []model.TLSConnection{}, "[api] Cannot read stats response"},
		{"parse", []model.TLSConnection{}, "[api] Cannot parse stats response"},
	}

	mockHttpClient := &mockHttpClient{}
	mockK8SClient := &mockK8SClient{}

	service := TlsParserService{repo: &repository.DbRepository{}, updater: &update.CertificateUpdater{}, httpClient: mockHttpClient, k8sClient: mockK8SClient}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			url := fmt.Sprintf("http://%%s:6676/tlsparser/connections/?scenario=%s", test.scenario)

			result, _ := service.buildConnectionsResponse(url)

			assert.EqualValues(t, test.want, result)
			assert.Contains(t, str.String(), test.error)

		})
	}

}

func TestBuildDetailsResponse(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		scenario string
		want     model.TLSDetails
		error    string
	}{
		{"ok_detail", dbDetails, ""},
		{"ok_detail_empty", model.TLSDetails{}, ""},
		{"error", model.TLSDetails{}, "[api] Cannot get stats"},
		{"read", model.TLSDetails{}, "[api] Cannot read stats response"},
		{"parse", model.TLSDetails{}, "[api] Cannot parse stats response"},
	}

	mockHttpClient := &mockHttpClient{}
	mockK8SClient := &mockK8SClient{}

	service := TlsParserService{repo: &repository.DbRepository{}, updater: &update.CertificateUpdater{}, httpClient: mockHttpClient, k8sClient: mockK8SClient}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			url := fmt.Sprintf("http://%%s:6676/tlsparser/connections/%s?scenario=%s", "id1", test.scenario)

			result, _ := service.buildDetailsResponse(url)

			assert.EqualValues(t, test.want, result)
			assert.Contains(t, str.String(), test.error)

		})
	}

}
