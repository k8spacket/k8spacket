package db

import (
	tcp_model "github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	tls_model "github.com/k8spacket/k8spacket/internal/modules/tls-parser/model"
	"github.com/timshannon/bolthold"
)

type IDBHandler[T tls_model.TLSDetails | tls_model.TLSConnection | tcp_model.ConnectionItem] interface {
	Query(query *bolthold.Query) ([]T, error)
	QueryMatchFunc(field string, matchFunc func(*T) (bool, error)) bolthold.Query
	Read(key string) (T, error)
	Upsert(key string, value *T) error
	Close() error
}
