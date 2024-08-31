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

	"github.com/k8spacket/k8spacket/external/handlerio"
	httpclient "github.com/k8spacket/k8spacket/external/http"
	k8sclient "github.com/k8spacket/k8spacket/external/k8s"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/modules/nodegraph/stats"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.ConnectionItem{
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour * -1), Src: "test", ConnCount: 10, ConnPersistent: 3, MaxDuration: 1},
	model.ConnectionItem{LastSeen: time.Now(), SrcNamespace: "test", SrcName: "test", ConnCount: 4, ConnPersistent: 0},
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour), DstNamespace: "test", Dst: "test", ConnCount: 101, ConnPersistent: 77},
}

type mockRepository struct {
	repo   repository.IRepository[model.ConnectionItem]
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
	k8sClient k8sclient.IK8SClient
}

func (k8sClient *mockK8SClient) GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string {
	return []string{"127.0.0.1"}
}

type mockHttpClient struct {
	httpClient httpclient.IHttpClient
}

func (httpClient *mockHttpClient) Do(req *http.Request) (*http.Response, error) {

	if req.URL.Query().Get("scenario") == "ok" {
		result, _ := json.Marshal(dbState)
		return &http.Response{
			Body: io.NopCloser(bytes.NewBuffer(result)),
		}, nil
	}
	if req.URL.Query().Get("scenario") == "error" {
		response := []model.ConnectionItem{}
		err := errors.New("error")
		result, _ := json.Marshal(response)
		return &http.Response{
			Body: io.NopCloser(bytes.NewBuffer(result)),
		}, err
	}
	if req.URL.Query().Get("scenario") == "read" {
		reader := BrokenReader{}
		return &http.Response{
			Body: &reader,
		}, nil
	}
	if req.URL.Query().Get("scenario") == "parse" {
		result := []byte("parse error")
		return &http.Response{
			Body: io.NopCloser(bytes.NewBuffer(result)),
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
	handlerio.IHandlerIO
}

func (mockHandlerIO *mockHandlerIO) ReadFile(name string) ([]byte, error) {
	if mockHandlerIO.scenario == "error" {
		return []byte{}, errors.New("error")
	}
	return os.ReadFile("../../fields.json")
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
	service := &Service{mockRepository, &stats.Factory{}, &httpclient.HttpClient{}, &k8sclient.K8SClient{}, &handlerio.HandlerIO{}}

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
			service := &Service{mockRepository, &stats.Factory{}, &httpclient.HttpClient{}, &k8sclient.K8SClient{}, &handlerio.HandlerIO{}}

			service.update("src", "srcName", "srcNs", "dst", "dstName", "dstNs", true, 100, 200, 1)

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
				model.Node{Id: "test", Title: "", SubTitle: "test", MainStat: "all: 101", SecondaryStat: "persistent: 77", Arc1: 0.7623762376237624, Arc2: 0.2376237623762376, Arc3: 0},
				model.Node{Id: "", Title: "", SubTitle: "", MainStat: "all: 14", SecondaryStat: "persistent: 3", Arc1: 0.21428571428571427, Arc2: 0.7857142857142857, Arc3: 0},
				model.Node{Id: "", Title: "", SubTitle: "", MainStat: "all: 14", SecondaryStat: "persistent: 3", Arc1: 0.21428571428571427, Arc2: 0.7857142857142857, Arc3: 0},
				model.Node{Id: "", Title: "", SubTitle: "", MainStat: "all: 14", SecondaryStat: "persistent: 3", Arc1: 0.21428571428571427, Arc2: 0.7857142857142857, Arc3: 0},
				model.Node{Id: "", Title: "", SubTitle: "", MainStat: "all: 14", SecondaryStat: "persistent: 3", Arc1: 0.21428571428571427, Arc2: 0.7857142857142857, Arc3: 0},
				model.Node{Id: "test", Title: "", SubTitle: "test", MainStat: "all: 101", SecondaryStat: "persistent: 77", Arc1: 0.7623762376237624, Arc2: 0.2376237623762376, Arc3: 0}},
			Edges: []model.Edge{
				model.Edge{Id: "test-", Source: "test", Target: "", MainStat: "all: 10", SecondaryStat: "persistent: 3"},
				model.Edge{Id: "-", Source: "", Target: "", MainStat: "all: 4", SecondaryStat: "persistent: 0"},
				model.Edge{Id: "-test", Source: "", Target: "test", MainStat: "all: 101", SecondaryStat: "persistent: 77"}}}, ""},
		{"error", &model.NodeGraph{}, "[api] Cannot get stats"},
		{"read", &model.NodeGraph{}, "[api] Cannot read stats response"},
		{"parse", &model.NodeGraph{}, "[api] Cannot parse stats response"},
	}

	mockRepository := &mockRepository{}
	service := &Service{mockRepository, &stats.Factory{}, &mockHttpClient{}, &mockK8SClient{}, &handlerio.HandlerIO{}}

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
			service := &Service{mockRepository, &stats.Factory{}, &mockHttpClient{}, &mockK8SClient{}, mockHandlerIO}

			result, _ := service.getO11yStatsConfig(test.scenario)

			resultStr := Fields{}
			json.Unmarshal([]byte(result), &resultStr)

			assert.EqualValues(t, test.want, resultStr)

			assert.Contains(t, str.String(), test.err)
		})
	}

}
