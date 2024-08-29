package nodegraph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"testing"
	"time"

	httpclient "github.com/k8spacket/k8spacket/external/http"
	k8sclient "github.com/k8spacket/k8spacket/external/k8s"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/modules/nodegraph/stats"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.ConnectionItem {
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour * -1), Src: "test"},
	model.ConnectionItem{LastSeen: time.Now(), SrcNamespace: "test", SrcName: "test"},
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour), DstNamespace: "test", Dst: "test"},
}

type mockRepository struct {
	repo repository.IRepository[model.ConnectionItem]
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


func TestGetConnections(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	mockRepository := &mockRepository{}
	factory := &stats.Factory{}
	service := &Service{mockRepository, factory, &httpclient.HttpClient{}, &k8sclient.K8SClient{}}

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
		want     model.ConnectionItem
	}{
		{model.ConnectionItem{Src: "src", Dst: "dst", ConnCount: 10, ConnPersistent: 5, BytesReceived: 1000, BytesSent: 500, Duration: 0.5, MaxDuration: 0.5}, 
			model.ConnectionItem{Src:"src", SrcName:"srcName", SrcNamespace:"srcNs", Dst:"dst", DstName:"dstName", DstNamespace:"dstNs", ConnCount:11, ConnPersistent:6, BytesSent:600, BytesReceived:1200, Duration:1.5, MaxDuration:1}},
		{model.ConnectionItem{},
			model.ConnectionItem{Src:"src", SrcName:"srcName", SrcNamespace:"srcNs", Dst:"dst", DstName:"dstName", DstNamespace:"dstNs", ConnCount:1, ConnPersistent:1, BytesSent:100, BytesReceived:200, Duration:1, MaxDuration:1}},
	}

	for _, test := range tests {
		t.Run(test.item.Src, func(t *testing.T) {

		mockRepository := &mockRepository{result: test.item}
		factory := &stats.Factory{}
		service := &Service{mockRepository, factory, &httpclient.HttpClient{}, &k8sclient.K8SClient{}}

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
		Nodes:[]model.Node{
			model.Node{Id:"", Title:"test", SubTitle:"", MainStat:"all: N/A", SecondaryStat:"persistent: N/A", Arc1:0, Arc2:0, Arc3:0}, 
			model.Node{Id:"test", Title:"", SubTitle:"test", MainStat:"all: N/A", SecondaryStat:"persistent: N/A", Arc1:0, Arc2:0, Arc3:0}, 
			model.Node{Id:"test", Title:"", SubTitle:"test", MainStat:"all: N/A", SecondaryStat:"persistent: N/A", Arc1:0, Arc2:0, Arc3:0}, 
			model.Node{Id:"", Title:"test", SubTitle:"", MainStat:"all: N/A", SecondaryStat:"persistent: N/A", Arc1:0, Arc2:0, Arc3:0}, 
			model.Node{Id:"", Title:"test", SubTitle:"", MainStat:"all: N/A", SecondaryStat:"persistent: N/A", Arc1:0, Arc2:0, Arc3:0}, 
			model.Node{Id:"", Title:"test", SubTitle:"", MainStat:"all: N/A", SecondaryStat:"persistent: N/A", Arc1:0, Arc2:0, Arc3:0}}, 
		Edges:[]model.Edge{
			model.Edge{Id:"-test", Source:"", Target:"test", MainStat:"all: 0", SecondaryStat:"persistent: 0"}, 
			model.Edge{Id:"test-", Source:"test", Target:"", MainStat:"all: 0", SecondaryStat:"persistent: 0"}, 
			model.Edge{Id:"-", Source:"", Target:"", MainStat:"all: 0", SecondaryStat:"persistent: 0"}}}, ""},
			{"error", &model.NodeGraph{}, "[api] Cannot get stats"},
			{"read", &model.NodeGraph{}, "[api] Cannot read stats response"},
			{"parse", &model.NodeGraph{}, "[api] Cannot parse stats response"},
	}

	mockRepository := &mockRepository{}
	factory := &stats.Factory{}
	service := &Service{mockRepository, factory, &mockHttpClient{}, &mockK8SClient{}}

	
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