package db

import (
	"fmt"
	tcp_model "github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	tls_model "github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/timshannon/bolthold"
	"go.etcd.io/bbolt"
	"hash/fnv"
)

type BoltDb[T tls_model.TLSDetails | tls_model.TLSConnection | tcp_model.ConnectionItem] struct {
	store *bolthold.Store
}

func New[T tls_model.TLSDetails | tls_model.TLSConnection | tcp_model.ConnectionItem](dbname string) (Db[T], error) {
	database, err := bolthold.Open(fmt.Sprintf("%s.db", dbname), 0600, nil)
	if err != nil {
		return nil, err
	}
	return &BoltDb[T]{database}, nil

}

func (boltDb *BoltDb[T]) Close() error {
	return boltDb.store.Close()
}

func (boltDb *BoltDb[T]) Read(key string) (T, error) {
	var value T
	return value, boltDb.store.Bolt().View(func(tx *bbolt.Tx) error {
		err := boltDb.store.TxGet(tx, key, &value)
		if err != nil {
			return err
		}
		return nil
	})
}

func (boltDb *BoltDb[T]) Query(query *bolthold.Query) ([]T, error) {
	var value []T
	return value, boltDb.store.Bolt().View(func(tx *bbolt.Tx) error {
		err := boltDb.store.TxFind(tx, &value, query)
		if err != nil {
			return err
		}
		return nil
	})
}

func (boltDb *BoltDb[T]) QueryMatchFunc(field string, matchFunc func(*T) (bool, error)) bolthold.Query {
	return *bolthold.Where(field).MatchFunc(func(ra *bolthold.RecordAccess) (bool, error) {
		record := ra.Record().(*T)
		return matchFunc(record)
	})
}

func (boltDb *BoltDb[T]) Upsert(key string, value *T) error {
	return boltDb.store.Bolt().Update(
		func(tx *bbolt.Tx) error {
			return boltDb.store.TxUpsert(tx, key, value)
		})
}

func HashId(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
