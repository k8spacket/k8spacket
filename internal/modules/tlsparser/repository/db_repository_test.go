package repository

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
	"github.com/stretchr/testify/assert"
	"github.com/timshannon/bolthold"
)

var dbState = []model.TLSConnection{
	{Src: "past", LastSeen: time.Now().Add(time.Hour * -1)},
	{Src: "now", LastSeen: time.Now()},
	{Src: "future", LastSeen: time.Now().Add(time.Hour * 1)},
	{Src: "error", LastSeen: time.Now().Add(time.Hour * 1000)},
}

type mockConnectionDb struct {
	Db          db.Db[model.TLSConnection]
	queryResult []model.TLSConnection
}

type mockDetailsDb struct {
	DBHandler db.Db[model.TLSDetails]
	fnCalled  bool
}

func (mock *mockConnectionDb) Query(query *bolthold.Query) ([]model.TLSConnection, error) {
	if mock.queryResult[0].LastSeen.After(time.Now().Add(time.Hour * 999)) {
		return []model.TLSConnection{}, errors.New("error")
	}
	return mock.queryResult, nil
}

func (mock *mockConnectionDb) QueryMatchFunc(field string, matchFunc func(*model.TLSConnection) (bool, error)) bolthold.Query {
	mock.queryResult = []model.TLSConnection{}
	for _, item := range dbState {
		matched, _ := matchFunc(&item)
		if matched {
			mock.queryResult = append(mock.queryResult, item)
		}
	}

	return bolthold.Query{}
}

func (mock *mockConnectionDb) Close() error {
	return nil
}

func (mock *mockConnectionDb) Read(key string) (model.TLSConnection, error) {
	return model.TLSConnection{}, nil
}

func (mock *mockConnectionDb) Upsert(key string, value *model.TLSConnection) error {
	if key == "error" {
		return errors.New("error")
	}
	value.UsedCipherSuite = value.UsedCipherSuite + "-TEST"
	return nil
}

func (mock *mockDetailsDb) Query(query *bolthold.Query) ([]model.TLSDetails, error) {
	return []model.TLSDetails{}, nil
}

func (mock *mockDetailsDb) QueryMatchFunc(field string, matchFunc func(*model.TLSDetails) (bool, error)) bolthold.Query {
	return bolthold.Query{}
}

func (mock *mockDetailsDb) Close() error {
	return nil
}

func (mock *mockDetailsDb) Read(key string) (model.TLSDetails, error) {
	if key == "error" {
		return model.TLSDetails{}, errors.New("cannot read db")
	}
	return model.TLSDetails{Domain: "test.com", UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, nil
}

func (mock *mockDetailsDb) Upsert(key string, value *model.TLSDetails) error {
	if key == "error" {
		return errors.New("error")
	}
	value.UsedCipherSuite = value.UsedCipherSuite + "-TEST"
	return nil
}

func TestQuery(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		msg      string
		from, to time.Time
		want     []model.TLSConnection
		error    string
	}{
		{"from / to filter", time.Now().Add(time.Minute * -1), time.Now().Add(time.Minute), dbState[1:2], ""},
		{"error", time.Now().Add(time.Hour * 998), time.Now().Add(time.Hour * 1001), []model.TLSConnection{}, "[db:tls_connections:Query] Error=error"},
	}

	mockConnectionDBHandler := &mockConnectionDb{}
	mockDetailsDBHandler := &mockDetailsDb{}

	repository := DbRepository{mockConnectionDBHandler, mockDetailsDBHandler}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {

			result := repository.Query(test.from, test.to)

			assert.EqualValues(t, test.want, result)
			assert.Contains(t, str.String(), test.error)
		})
	}

}

func TestRead(t *testing.T) {

	var tests = []struct {
		key  string
		want model.TLSDetails
	}{
		{"key", model.TLSDetails{Domain: "test.com", UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}},
		{"error", model.TLSDetails{}},
	}

	mockConnectionDBHandler := &mockConnectionDb{}
	mockDetailsDBHandler := &mockDetailsDb{}

	repository := DbRepository{mockConnectionDBHandler, mockDetailsDBHandler}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			t.Parallel()

			item := repository.Read(test.key)

			assert.EqualValues(t, item, test.want)
		})
	}
}

func TestUpsertConnection(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		key        string
		item, want model.TLSConnection
		error      string
	}{
		{"key", model.TLSConnection{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, model.TLSConnection{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384-TEST"}, ""},
		{"error", model.TLSConnection{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, model.TLSConnection{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, "[db:tls_connections:Upsert] Error=error"},
	}

	mockConnectionDBHandler := &mockConnectionDb{}
	mockDetailsDBHandler := &mockDetailsDb{}

	repository := DbRepository{mockConnectionDBHandler, mockDetailsDBHandler}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			t.Parallel()

			repository.UpsertConnection(test.key, &test.item)

			assert.EqualValues(t, test.want, test.item)
			assert.Contains(t, str.String(), test.error)

		})
	}
}

func TestUpsertDetails(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	var tests = []struct {
		key        string
		item, want model.TLSDetails
		error      string
	}{
		{"key", model.TLSDetails{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, model.TLSDetails{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384-TEST"}, ""},
		{"error", model.TLSDetails{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, model.TLSDetails{UsedCipherSuite: "ECDHE-RSA-AES256-GCM-SHA384"}, "[db:tls_details:Upsert] Error=error"},
	}

	mockConnectionDBHandler := &mockConnectionDb{}
	mockDetailsDBHandler := &mockDetailsDb{}

	repository := DbRepository{mockConnectionDBHandler, mockDetailsDBHandler}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			t.Parallel()

			repository.UpsertDetails(test.key, &test.item, func(newValue, oldValue *model.TLSDetails) {
				mockDetailsDBHandler.fnCalled = true
			})

			assert.EqualValues(t, test.want, test.item)
			assert.EqualValues(t, true, mockDetailsDBHandler.fnCalled)
			assert.Contains(t, str.String(), test.error)

		})
	}
}
