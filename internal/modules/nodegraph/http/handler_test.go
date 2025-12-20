package http

import (
	"encoding/json"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/repository"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

var dbState = []model.ConnectionItem{
	{LastSeen: time.Now().Add(time.Hour * -1).UTC(), Src: "client-1", Dst: "server-1", ConnCount: 10, ConnPersistent: 3, MaxDuration: 1},
	{LastSeen: time.Now().UTC(), Src: "client-1", Dst: "server-2", SrcNamespace: "test", SrcName: "test", ConnCount: 6, ConnPersistent: 4},
	{LastSeen: time.Now().Add(time.Hour).UTC(), Src: "client-2", Dst: "server-2", DstNamespace: "test", ConnCount: 4, ConnPersistent: 0},
	{LastSeen: time.Now().Add(time.Hour).UTC(), Src: "client-3", Dst: "server-3", DstNamespace: "test", ConnCount: 101, ConnPersistent: 77},
}

type mockRepository struct {
	repository.Repository[model.ConnectionItem]
	from, to                        time.Time
	patternNs, patternIn, patternEx string
	client, server                  string
}

func (mock *mockRepository) Query(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {
	mock.from = from
	mock.to = to
	mock.patternNs = patternNs.String()
	mock.patternIn = patternIn.String()
	mock.patternEx = patternEx.String()
	return dbState
}

func TestConnectionHandler(t *testing.T) {

	mockRepository := &mockRepository{}
	handler := NewHandler(mockRepository)

	req, err := http.NewRequest("GET", "/nodegraph/connections", nil)
	if err != nil {
		t.Fatal(err)
	}

	from := int64(1609506000000)
	fromTime := time.Unix(from/1000, 0)

	to := int64(1609506000000)
	toTime := time.Unix(to/1000, 0)

	q := req.URL.Query()
	q.Add("from", strconv.FormatInt(from, 10))
	q.Add("to", strconv.FormatInt(to, 10))
	q.Add("namespace", "ns")
	q.Add("include", "in")
	q.Add("exclude", "ex")
	req.URL.RawQuery = q.Encode()

	rr := httptest.NewRecorder()
	httpHandler := http.HandlerFunc(handler.ConnectionHandler)

	httpHandler.ServeHTTP(rr, req)

	assert.EqualValues(t, rr.Code, http.StatusOK)

	var response []model.ConnectionItem
	json.Unmarshal([]byte(rr.Body.String()), &response)

	assert.EqualValues(t, dbState, response)

	assert.EqualValues(t, fromTime, mockRepository.from)
	assert.EqualValues(t, toTime, mockRepository.to)
	assert.EqualValues(t, "ns", mockRepository.patternNs)
	assert.EqualValues(t, "in", mockRepository.patternIn)
	assert.EqualValues(t, "ex", mockRepository.patternEx)

}
