package repository

import (
	"bytes"
	"errors"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/modules/db"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
	"github.com/timshannon/bolthold"
)

var dbState = []model.ConnectionItem{
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour * -1), Src: "test"},
	model.ConnectionItem{LastSeen: time.Now(), SrcNamespace: "test", SrcName: "test"},
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour), DstNamespace: "test", Dst: "test"},
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour * 2), DstName: "test"},
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour * 2)},
	model.ConnectionItem{LastSeen: time.Now().Add(time.Hour * 1000)},
}

type mockDBHandler struct {
	DBHandler   db.IDBHandler[model.ConnectionItem]
	queryResult []model.ConnectionItem
}

func (mock *mockDBHandler) Read(key string) (model.ConnectionItem, error) {
	if key == "error" {
		return model.ConnectionItem{}, errors.New("cannot read db")
	}
	return model.ConnectionItem{BytesSent: 300, BytesReceived: 100, Duration: 0.5}, nil
}

func (mock *mockDBHandler) Close() error {
	return nil
}

func (mock *mockDBHandler) Query(query *bolthold.Query) ([]model.ConnectionItem, error) {
	if mock.queryResult[0].LastSeen.After(time.Now().Add(time.Hour * 999)) {
		return []model.ConnectionItem{}, errors.New("error")
	}
	return mock.queryResult, nil
}

func (mock *mockDBHandler) QueryMatchFunc(field string, matchFunc func(*model.ConnectionItem) (bool, error)) bolthold.Query {
	mock.queryResult = []model.ConnectionItem{}
	for _, item := range dbState {
		matched, _ := matchFunc(&item)
		if matched {
			mock.queryResult = append(mock.queryResult, item)
		}
	}

	return bolthold.Query{}
}

func (mock *mockDBHandler) Upsert(key string, value *model.ConnectionItem) error {
	if key == "error" {
		return errors.New("error")
	}
	value.BytesReceived++
	return nil
}

func TestRead(t *testing.T) {

	var tests = []struct {
		key  string
		want model.ConnectionItem
	}{
		{"key", model.ConnectionItem{BytesSent: 300, BytesReceived: 100, Duration: 0.5}},
		{"error", model.ConnectionItem{}},
	}

	mockDBHandler := &mockDBHandler{}

	repository := Repository{mockDBHandler}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			t.Parallel()

			connectionItem := repository.Read(test.key)

			assert.EqualValues(t, connectionItem, test.want)
		})
	}
}

func TestQuery(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		msg                             string
		from, to                        time.Time
		patternNs, patternIn, patternEx *regexp.Regexp
		want                            []model.ConnectionItem
		error                           string
	}{
		{"from / to filter", time.Now().Add(time.Minute * -1), time.Now().Add(time.Minute), regexp.MustCompile(""), regexp.MustCompile(""), regexp.MustCompile(""), dbState[1:2], ""},
		{"namespace filter", time.Now().Add(time.Hour * -3), time.Now().Add(time.Hour * 3), regexp.MustCompile("^test$"), regexp.MustCompile(""), regexp.MustCompile(""), dbState[1:3], ""},
		{"include filter", time.Now().Add(time.Hour * -3), time.Now().Add(time.Hour * 3), regexp.MustCompile(""), regexp.MustCompile("test"), regexp.MustCompile(""), dbState[0:4], ""},
		{"exclude filter", time.Now().Add(time.Hour * -3), time.Now().Add(time.Hour * 3), regexp.MustCompile(""), regexp.MustCompile(""), regexp.MustCompile("test"), dbState[4:5], ""},
		{"error", time.Now().Add(time.Hour * 998), time.Now().Add(time.Hour * 1001), regexp.MustCompile(""), regexp.MustCompile(""), regexp.MustCompile(""), []model.ConnectionItem{}, "[db:tcp_connections:Query] Error=error"},
	}

	mockDBHandler := &mockDBHandler{}

	repository := Repository{mockDBHandler}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {

			result := repository.Query(test.from, test.to, test.patternNs, test.patternIn, test.patternEx)

			assert.EqualValues(t, test.want, result)
			assert.Contains(t, str.String(), test.error)

		})
	}
}

func TestSet(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		key        string
		item, want model.ConnectionItem
		error      string
	}{
		{"key", model.ConnectionItem{BytesReceived: 100}, model.ConnectionItem{BytesReceived: 101}, ""},
		{"error", model.ConnectionItem{BytesReceived: 666}, model.ConnectionItem{BytesReceived: 666}, "[db:tcp_connections:Upsert] Error=error"},
	}

	mockDBHandler := &mockDBHandler{}

	repository := Repository{mockDBHandler}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			t.Parallel()

			repository.Set(test.key, &test.item)

			assert.EqualValues(t, test.want, test.item)
			assert.Contains(t, str.String(), test.error)

		})
	}
}
