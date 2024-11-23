package nodegraph

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

var repo = []model.ConnectionItem{
	{Src: "src", Dst: "dst"},
	{Src: "src1", Dst: "dst1"},
}

type mockService struct {
	IService
	from, to                        time.Time
	patternNs, patternIn, patternEx string
	client, server                  string
}

func (mockService *mockService) getConnections(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {
	mockService.from = from
	mockService.to = to
	mockService.patternNs = patternNs.String()
	mockService.patternIn = patternIn.String()
	mockService.patternEx = patternEx.String()
	return repo
}

func TestConnectionHandler(t *testing.T) {

	service := &mockService{}
	controller := &Controller{service: service}

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
	handler := http.HandlerFunc(controller.ConnectionHandler)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(t, rr.Code, http.StatusOK)

	var response []model.ConnectionItem
	json.Unmarshal([]byte(rr.Body.String()), &response)

	assert.EqualValues(t, repo, response)

	assert.EqualValues(t, fromTime, service.from)
	assert.EqualValues(t, toTime, service.to)
	assert.EqualValues(t, "ns", service.patternNs)
	assert.EqualValues(t, "in", service.patternIn)
	assert.EqualValues(t, "ex", service.patternEx)

}
