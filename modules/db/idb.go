package db

import (
	tcp_model "github.com/k8spacket/k8spacket/modules/nodegraph/model"
	tls_model "github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"github.com/timshannon/bolthold"
)

type IDBHandler[T tls_model.TLSDetails | tls_model.TLSConnection | tcp_model.ConnectionItem] interface {
	Close() error
	Read(key string) (T, error)
	Query(query *bolthold.Query) ([]T, error)
	Upsert(key string, value T) error
}
