package idb

import (
	"fmt"
	tcp_model "github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph/model"
	tls_model "github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
	"github.com/timshannon/bolthold"
	"go.etcd.io/bbolt"
	"hash/fnv"
)

type DB[T tls_model.TLSDetails | tls_model.TLSConnection | tcp_model.ConnectionItem] struct {
	store *bolthold.Store
}

func StartDB[T tls_model.TLSDetails | tls_model.TLSConnection | tcp_model.ConnectionItem](dbname string) (*DB[T], error) {

	database, err := bolthold.Open(fmt.Sprintf("%s.db", dbname), 0600, nil)
	if err != nil {
		return nil, err
	}
	return &DB[T]{database}, nil
}

func (k *DB[T]) Close() error {
	return k.store.Close()
}

func (k *DB[T]) Read(key string) (T, error) {
	var value T
	return value, k.store.Bolt().View(func(tx *bbolt.Tx) error {
		err := k.store.TxGet(tx, key, &value)
		if err != nil {
			return err
		}
		return nil
	})
}

func (k *DB[T]) Query(query *bolthold.Query) ([]T, error) {
	var value []T
	return value, k.store.Bolt().View(func(tx *bbolt.Tx) error {
		err := k.store.TxFind(tx, &value, query)
		if err != nil {
			return err
		}
		return nil
	})
}

func (k *DB[T]) Upsert(key string, value T) error {
	return k.store.Bolt().Update(
		func(tx *bbolt.Tx) error {
			return k.store.TxUpsert(tx, key, value)
		})
}

func HashId(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
