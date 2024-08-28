package repository

import (
	"errors"
	"testing"

	"github.com/k8spacket/k8spacket/modules/db"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
	"github.com/timshannon/bolthold"
)

type mockDBHandler struct {
	DBHandler db.IDBHandler[model.ConnectionItem]
}

func (mock *mockDBHandler) Read(key string) (model.ConnectionItem, error) {
	if(key == "error") {
		return model.ConnectionItem{}, errors.New("cannot read db")
	}
	return model.ConnectionItem{BytesSent: 300, BytesReceived: 100, Duration: 0.5}, nil
}

func (mock *mockDBHandler) Close() error {
	return nil
}


func (mock *mockDBHandler) Query(query *bolthold.Query) ([]model.ConnectionItem, error) {
	return []model.ConnectionItem{}, nil
}

func (mock *mockDBHandler) QueryMatchFunc(field string, matchFunc func(*model.ConnectionItem) (bool, error)) bolthold.Query {
	return bolthold.Query{}
}

func (mock *mockDBHandler) Upsert(key string, value model.ConnectionItem) error {
	return nil
}

func TestRead(t *testing.T) {

	var tests = []struct {
		key string;
		want model.ConnectionItem
	} {
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
