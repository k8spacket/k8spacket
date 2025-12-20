package o11y

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/stats"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/k8spacket/k8spacket/internal/thirdparty/resource"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

type Field struct {
	FieldName   string `json:"field_name"`
	Type        string `json:"type"`
	Color       string `json:"color"`
	DisplayName string `json:"displayName"`
}

type Fields struct {
	EdgesFields []Field `json:"edges_fields"`
	NodesFields []Field `json:"nodes_fields"`
}

var dbState = []model.ConnectionItem{
	{LastSeen: time.Now().Add(time.Hour * -1), Src: "client-1", Dst: "server-1", SrcNamespace: "test", SrcName: "test", ConnCount: 10, ConnPersistent: 3, MaxDuration: 1},
	{LastSeen: time.Now(), Src: "client-1", Dst: "server-2", SrcNamespace: "test", SrcName: "test", ConnCount: 6, ConnPersistent: 4},
	{LastSeen: time.Now().Add(time.Hour), Src: "client-2", Dst: "server-2", DstNamespace: "test", ConnCount: 4, ConnPersistent: 0},
	{LastSeen: time.Now().Add(time.Hour), Src: "client-3", Dst: "server-3", DstNamespace: "test", ConnCount: 101, ConnPersistent: 77},
}

type mockResource struct {
	resource.Resource
	scenario string
}

func (mockHandlerIO *mockResource) Read(name string) ([]byte, error) {
	if mockHandlerIO.scenario == "error" {
		return []byte{}, errors.New("error")
	}
	return os.ReadFile("../../../../fields.json")
}

type mockK8SClient struct {
	k8sclient.Client
}

func (k8sClient *mockK8SClient) GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string {
	return []string{"127.0.0.1"}
}

type mockHttpClient struct {
	httpclient.Client
	scenario string
}

func (mockHttpClient *mockHttpClient) Do(req *http.Request) (*http.Response, error) {

	if mockHttpClient.scenario == "ok" {
		result, _ := json.Marshal(dbState)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusOK,
		}, nil
	}
	if mockHttpClient.scenario == "error" {
		response := []model.ConnectionItem{}
		err := errors.New("error")
		result, _ := json.Marshal(response)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(result)),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	if mockHttpClient.scenario == "read" {
		reader := BrokenReader{}
		return &http.Response{
			Body:       &reader,
			StatusCode: http.StatusOK,
		}, nil
	}
	if mockHttpClient.scenario == "parse" {
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

func TestHealth(t *testing.T) {

	o11yController := NewO11yHandler(nil, nil, nil, nil)

	req, err := http.NewRequest("GET", "/nodegraph/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(o11yController.Health)
	handler.ServeHTTP(rr, req)

	assert.EqualValues(t, rr.Code, http.StatusOK)

}

func TestNodeGraphFieldsHandler(t *testing.T) {

	var tests = []struct {
		scenario string
		want     Fields
		status   int
		err      string
	}{
		{"connection", Fields{
			EdgesFields: []Field{
				{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "source", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "target", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "mainStat", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: ""}},
			NodesFields: []Field{
				{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "title", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "subTitle", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "mainStat", Type: "string", Color: "", DisplayName: "All connections "},
				{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: "Persistent connections "},
				{FieldName: "arc__1", Type: "number", Color: "green", DisplayName: "Persistent connections"},
				{FieldName: "arc__2", Type: "number", Color: "red", DisplayName: "Short-lived connections"}}}, http.StatusOK, ""},
		{"bytes", Fields{
			EdgesFields: []Field{
				{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "source", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "target", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "mainStat", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: ""}},
			NodesFields: []Field{
				{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "title", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "subTitle", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "mainStat", Type: "string", Color: "", DisplayName: "Bytes received "},
				{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: "Bytes responded "},
				{FieldName: "arc__1", Type: "number", Color: "blue", DisplayName: "Bytes received"},
				{FieldName: "arc__2", Type: "number", Color: "yellow", DisplayName: "Bytes responded"}}}, http.StatusOK, ""},
		{"duration", Fields{
			EdgesFields: []Field{
				{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "source", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "target", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "mainStat", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: ""}},
			NodesFields: []Field{
				{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "title", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "subTitle", Type: "string", Color: "", DisplayName: ""},
				{FieldName: "mainStat", Type: "string", Color: "", DisplayName: "Average duration "},
				{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: "Max duration "},
				{FieldName: "arc__1", Type: "number", Color: "purple", DisplayName: "Average duration"},
				{FieldName: "arc__2", Type: "number", Color: "white", DisplayName: "Max duration"}}}, http.StatusOK, ""},
		{"error", Fields{}, http.StatusInternalServerError, "error"},
	}

	mockResource := &mockResource{}
	o11yController := NewO11yHandler(&stats.StatsFactory{}, nil, nil, mockResource)

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()

			mockResource.scenario = test.scenario
			req, err := http.NewRequest("GET", "/nodegraph/api/graph/fields", nil)
			if err != nil {
				t.Fatal(err)
			}
			q := req.URL.Query()
			q.Add("stats-type", test.scenario)
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o11yController.NodeGraphFieldsHandler)
			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, test.status, rr.Code)

			var response Fields
			json.Unmarshal([]byte(rr.Body.String()), &response)
			assert.EqualValues(t, test.want, response)
		})
	}
}

//func TestNodeGraphDataHandler(t *testing.T) {
//
//	var tests = []struct {
//		scenario string
//		want     model.NodeGraph
//		status   int
//		err      string
//	}{
//		{"ok", model.NodeGraph{
//			Nodes: []model.Node{
//				{Id: "client-1", Title: "test", SubTitle: "client-1", MainStat: "all: N/A", SecondaryStat: "persistent: N/A", Arc1: 0, Arc2: 0, Arc3: 0},
//				{Id: "server-1", Title: "", SubTitle: "server-1", MainStat: "all: 10", SecondaryStat: "persistent: 3", Arc1: 0.3, Arc2: 0.7, Arc3: 0},
//				{Id: "server-2", Title: "", SubTitle: "server-2", MainStat: "all: 10", SecondaryStat: "persistent: 4", Arc1: 0.4, Arc2: 0.6, Arc3: 0},
//				{Id: "client-2", Title: "", SubTitle: "client-2", MainStat: "all: N/A", SecondaryStat: "persistent: N/A", Arc1: 0, Arc2: 0, Arc3: 0},
//				{Id: "client-3", Title: "", SubTitle: "client-3", MainStat: "all: N/A", SecondaryStat: "persistent: N/A", Arc1: 0, Arc2: 0, Arc3: 0},
//				{Id: "server-3", Title: "", SubTitle: "server-3", MainStat: "all: 101", SecondaryStat: "persistent: 77", Arc1: 0.7623762376237624, Arc2: 0.2376237623762376, Arc3: 0}},
//			Edges: []model.Edge{
//				{Id: "client-1-server-1", Source: "client-1", Target: "server-1", MainStat: "all: 10", SecondaryStat: "persistent: 3"},
//				{Id: "client-1-server-2", Source: "client-1", Target: "server-2", MainStat: "all: 6", SecondaryStat: "persistent: 4"},
//				{Id: "client-2-server-2", Source: "client-2", Target: "server-2", MainStat: "all: 4", SecondaryStat: "persistent: 0"},
//				{Id: "client-3-server-3", Source: "client-3", Target: "server-3", MainStat: "all: 101", SecondaryStat: "persistent: 77"}}}, http.StatusOK, ""},
//		{"error", model.NodeGraph{}, http.StatusInternalServerError, "error"},
//	}
//
//	mockResource := &mockResource{}
//	mockHttpClient := &mockHttpClient{}
//	mockK8SClient := &mockK8SClient{}
//	o11yController := NewO11yHandler(&stats.StatsFactory{}, mockHttpClient, mockK8SClient, mockResource)
//
//	for _, test := range tests {
//		t.Run(test.scenario, func(t *testing.T) {
//			t.Parallel()
//
//			mockHttpClient.scenario = test.scenario
//			req, err := http.NewRequest("GET", "/nodegraph/api/graph/data", nil)
//			if err != nil {
//				t.Fatal(err)
//			}
//			req.Header.Set("scenario", test.scenario)
//
//			rr := httptest.NewRecorder()
//			handler := http.HandlerFunc(o11yController.NodeGraphDataHandler)
//			handler.ServeHTTP(rr, req)
//
//			assert.EqualValues(t, test.status, rr.Code)
//
//			var resultGraph model.NodeGraph
//			json.Unmarshal([]byte(rr.Body.String()), &resultGraph)
//
//			assert.EqualValues(t, test.want.Nodes, resultGraph.Nodes)
//			assert.EqualValues(t, test.want.Edges, resultGraph.Edges)
//		})
//	}
//}
