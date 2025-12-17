package nodegraph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/stats"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/k8spacket/k8spacket/internal/thirdparty/resource"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.ConnectionItem{
	{LastSeen: time.Now().Add(time.Hour * -1), Src: "client-1", Dst: "server-1", ConnCount: 10, ConnPersistent: 3, MaxDuration: 1},
	{LastSeen: time.Now(), Src: "client-1", Dst: "server-2", SrcNamespace: "test", SrcName: "test", ConnCount: 6, ConnPersistent: 4},
	{LastSeen: time.Now().Add(time.Hour), Src: "client-2", Dst: "server-2", DstNamespace: "test", ConnCount: 4, ConnPersistent: 0},
	{LastSeen: time.Now().Add(time.Hour), Src: "client-3", Dst: "server-3", DstNamespace: "test", ConnCount: 101, ConnPersistent: 77},
}

type mockRepository struct {
	repo   repository.Repository[model.ConnectionItem]
	result model.ConnectionItem
}

func (mock *mockRepository) Query(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {
	return dbState
}

func (mock *mockRepository) Read(key string) model.ConnectionItem {
	return mock.result
}

func (mock *mockRepository) Set(key string, value *model.ConnectionItem) {
	mock.result = *value
}

type mockK8SClient struct {
	k8sClient k8sclient.Client
}

func (k8sClient *mockK8SClient) GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string {
	return []string{"127.0.0.1"}
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
	if req.URL.Query().Get("scenario") == "error" {
		response := []model.ConnectionItem{}
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

type mockHandlerIO struct {
	scenario string
	resource.Resource
}

func (mockHandlerIO *mockHandlerIO) Read(name string) ([]byte, error) {
	if mockHandlerIO.scenario == "error" {
		return []byte{}, errors.New("error")
	}
	return os.ReadFile("../../../fields.json")
}

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

func TestGetConnections(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	mockRepository := &mockRepository{}
	service := &NodegraphService{repo: mockRepository, factory: &stats.StatsFactory{}, httpClient: &httpclient.HttpClient{}, k8sClient: &k8sclient.K8SClient{}, resource: &resource.FileResource{}}

	from := time.Now().Add(time.Hour * -1)
	to := time.Now().Add(time.Hour)
	patternNs := regexp.MustCompile("ns")
	patternIn := regexp.MustCompile("in")
	patternEx := regexp.MustCompile("ex")

	result := service.getConnections(from, to, patternNs, patternIn, patternEx)

	assert.EqualValues(t, dbState, result)
	assert.Contains(t, str.String(), fmt.Sprintf("[api:params] patternNs=%s patternIn=%s patternEx=%s from=\"%s\" to=\"%s\"\n",
		patternNs, patternIn, patternEx, from.Format(time.DateTime), to.Format(time.DateTime)))

}

func TestUpdate(t *testing.T) {
	var tests = []struct {
		item model.ConnectionItem
		want model.ConnectionItem
	}{
		{model.ConnectionItem{Src: "src", Dst: "dst", ConnCount: 10, ConnPersistent: 5, BytesReceived: 1000, BytesSent: 500, Duration: 0.5, MaxDuration: 0.5},
			model.ConnectionItem{Src: "src", SrcName: "srcName", SrcNamespace: "srcNs", Dst: "dst", DstName: "dstName", DstNamespace: "dstNs", ConnCount: 11, ConnPersistent: 6, BytesSent: 600, BytesReceived: 1200, Duration: 1.5, MaxDuration: 1}},
		{model.ConnectionItem{},
			model.ConnectionItem{Src: "src", SrcName: "srcName", SrcNamespace: "srcNs", Dst: "dst", DstName: "dstName", DstNamespace: "dstNs", ConnCount: 1, ConnPersistent: 1, BytesSent: 100, BytesReceived: 200, Duration: 1, MaxDuration: 1}},
	}

	for _, test := range tests {
		t.Run(test.item.Src, func(t *testing.T) {

			mockRepository := &mockRepository{result: test.item}
			service := &NodegraphService{repo: mockRepository, factory: &stats.StatsFactory{}, httpClient: &httpclient.HttpClient{}, k8sClient: &k8sclient.K8SClient{}, resource: &resource.FileResource{}}

			service.update("src", "srcName", "srcNs", "dst", "dstName", "dstNs", true, 100, 200, 1, true)

			result := mockRepository.Read("")

			test.want.LastSeen = result.LastSeen
			assert.EqualValues(t, test.want, result)
		})
	}
}

func TestBuildO11yResponse(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		scenario string
		want     *model.NodeGraph
		err      string
	}{

		{"ok", &model.NodeGraph{
			Nodes: []model.Node{
				{Id: "client-1", Title: "", SubTitle: "client-1", MainStat: "all: N/A", SecondaryStat: "persistent: N/A", Arc1: 0, Arc2: 0, Arc3: 0},
				{Id: "server-1", Title: "", SubTitle: "server-1", MainStat: "all: 10", SecondaryStat: "persistent: 3", Arc1: 0.3, Arc2: 0.7, Arc3: 0},
				{Id: "server-2", Title: "", SubTitle: "server-2", MainStat: "all: 10", SecondaryStat: "persistent: 4", Arc1: 0.4, Arc2: 0.6, Arc3: 0},
				{Id: "client-2", Title: "", SubTitle: "client-2", MainStat: "all: N/A", SecondaryStat: "persistent: N/A", Arc1: 0, Arc2: 0, Arc3: 0},
				{Id: "client-3", Title: "", SubTitle: "client-3", MainStat: "all: N/A", SecondaryStat: "persistent: N/A", Arc1: 0, Arc2: 0, Arc3: 0},
				{Id: "server-3", Title: "", SubTitle: "server-3", MainStat: "all: 101", SecondaryStat: "persistent: 77", Arc1: 0.7623762376237624, Arc2: 0.2376237623762376, Arc3: 0}},
			Edges: []model.Edge{
				{Id: "client-3-server-3", Source: "client-3", Target: "server-3", MainStat: "all: 101", SecondaryStat: "persistent: 77"},
				{Id: "client-1-server-1", Source: "client-1", Target: "server-1", MainStat: "all: 10", SecondaryStat: "persistent: 3"},
				{Id: "client-1-server-2", Source: "client-1", Target: "server-2", MainStat: "all: 6", SecondaryStat: "persistent: 4"},
				{Id: "client-2-server-2", Source: "client-2", Target: "server-2", MainStat: "all: 4", SecondaryStat: "persistent: 0"}}}, ""},
		{"error", &model.NodeGraph{}, "[api] Cannot get stats"},
		{"read", &model.NodeGraph{}, "[api] Cannot read stats response"},
		{"parse", &model.NodeGraph{}, "[api] Cannot parse stats response"},
	}

	mockRepository := &mockRepository{}
	service := &NodegraphService{repo: mockRepository, factory: &stats.StatsFactory{}, httpClient: &mockHttpClient{}, k8sClient: &mockK8SClient{}, resource: &resource.FileResource{}}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			r, _ := http.NewRequest(http.MethodGet, "", nil)

			q := r.URL.Query()
			q.Set("stats-type", "connection")
			q.Set("scenario", test.scenario)
			r.URL.RawQuery = q.Encode()

			result, _ := service.buildO11yResponse(r)

			assert.ElementsMatch(t, test.want.Nodes, result.Nodes)
			assert.ElementsMatch(t, test.want.Edges, result.Edges)

			assert.Contains(t, str.String(), test.err)
		})
	}
}

func TestGetO11yStatsConfig(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		scenario string
		want     Fields
		err      string
	}{
		{"connection", Fields{
			EdgesFields: []Field{
				Field{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "source", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "target", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "mainStat", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: ""}},
			NodesFields: []Field{
				Field{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "title", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "subTitle", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "mainStat", Type: "string", Color: "", DisplayName: "All connections "},
				Field{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: "Persistent connections "},
				Field{FieldName: "arc__1", Type: "number", Color: "green", DisplayName: "Persistent connections"},
				Field{FieldName: "arc__2", Type: "number", Color: "red", DisplayName: "Short-lived connections"}}}, ""},
		{"bytes", Fields{
			EdgesFields: []Field{
				Field{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "source", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "target", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "mainStat", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: ""}},
			NodesFields: []Field{
				Field{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "title", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "subTitle", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "mainStat", Type: "string", Color: "", DisplayName: "Bytes received "},
				Field{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: "Bytes responded "},
				Field{FieldName: "arc__1", Type: "number", Color: "blue", DisplayName: "Bytes received"},
				Field{FieldName: "arc__2", Type: "number", Color: "yellow", DisplayName: "Bytes responded"}}}, ""},
		{"duration", Fields{
			EdgesFields: []Field{
				Field{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "source", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "target", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "mainStat", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: ""}},
			NodesFields: []Field{
				Field{FieldName: "id", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "title", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "subTitle", Type: "string", Color: "", DisplayName: ""},
				Field{FieldName: "mainStat", Type: "string", Color: "", DisplayName: "Average duration "},
				Field{FieldName: "secondaryStat", Type: "string", Color: "", DisplayName: "Max duration "},
				Field{FieldName: "arc__1", Type: "number", Color: "purple", DisplayName: "Average duration"},
				Field{FieldName: "arc__2", Type: "number", Color: "white", DisplayName: "Max duration"}}}, ""},
		{"error", Fields{}, "\"Cannot read file\" Error=error"},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

			mockRepository := &mockRepository{}
			mockHandlerIO := &mockHandlerIO{scenario: test.scenario}
			service := &NodegraphService{repo: mockRepository, factory: &stats.StatsFactory{}, httpClient: &mockHttpClient{}, k8sClient: &mockK8SClient{}, resource: mockHandlerIO}

			result, _ := service.getO11yStatsConfig(test.scenario)

			resultStr := Fields{}
			json.Unmarshal([]byte(result), &resultStr)

			assert.EqualValues(t, test.want, resultStr)

			assert.Contains(t, str.String(), test.err)
		})
	}

}
