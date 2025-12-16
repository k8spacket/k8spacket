package nodegraph

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

var response = model.NodeGraph{
	Nodes: []model.Node{
		model.Node{Id: "test node"}},
	Edges: []model.Edge{
		model.Edge{Id: "test edge"}}}

func (mockNodegraphService *mockNodegraphService) getO11yStatsConfig(statsType string) (string, error) {
	if statsType == "error" {
		return "", errors.New("error")
	}
	return fmt.Sprintf("%s selected", statsType), nil
}

func (mockNodegraphService *mockNodegraphService) buildO11yResponse(r *http.Request) (model.NodeGraph, error) {
	if r.Header.Get("scenario") == "error" {
		return model.NodeGraph{}, errors.New("error")
	}
	return response, nil
}

func TestHealth(t *testing.T) {

	o11yController := &O11yController{service: &NodegraphService{}}

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
		want     string
		status   int
		err      string
	}{
		{"connection", "connection selected", http.StatusOK, ""},
		{"bytes", "bytes selected", http.StatusOK, ""},
		{"duration", "duration selected", http.StatusOK, ""},
		{"error", "error", http.StatusInternalServerError, "error"},
	}

	service := &mockNodegraphService{}
	o11yController := &O11yController{service: service}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()

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

			assert.EqualValues(t, rr.Code, test.status)
			assert.EqualValues(t, test.want, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func TestNodeGraphDataHandler(t *testing.T) {

	var tests = []struct {
		scenario string
		want     model.NodeGraph
		status   int
		err      string
	}{
		{"ok", response, http.StatusOK, ""},
		{"error", model.NodeGraph{}, http.StatusInternalServerError, "error"},
	}

	service := &mockNodegraphService{}
	o11yController := &O11yController{service: service}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("GET", "/nodegraph/api/graph/data", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("scenario", test.scenario)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o11yController.NodeGraphDataHandler)
			handler.ServeHTTP(rr, req)

			assert.EqualValues(t, rr.Code, test.status)

			var resultGraph model.NodeGraph
			json.Unmarshal([]byte(rr.Body.String()), &resultGraph)

			assert.EqualValues(t, test.want, resultGraph)
		})
	}
}
