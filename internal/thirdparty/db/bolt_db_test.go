package db

import (
	"path/filepath"
	"testing"

	tcp_model "github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
	"github.com/timshannon/bolthold"
)

func TestBoltDb_UpsertReadQuery(t *testing.T) {
	dbpath := filepath.Join(t.TempDir(), "testdb")
	db, err := New[tcp_model.ConnectionItem](dbpath)
	assert.NoError(t, err)
	defer db.Close()

	item := tcp_model.ConnectionItem{Src: "1.1.1.1", Dst: "2.2.2.2"}
	key := "k1"

	err = db.Upsert(key, &item)
	assert.NoError(t, err)

	got, err := db.Read(key)
	assert.NoError(t, err)
	assert.Equal(t, item.Src, got.Src)
	assert.Equal(t, item.Dst, got.Dst)

	// insert additional items for query
	item2 := tcp_model.ConnectionItem{Src: "3.3.3.3", Dst: "4.4.4.4"}
	_ = db.Upsert("k2", &item2)
	item3 := tcp_model.ConnectionItem{Src: "1.1.1.1", Dst: "5.5.5.5"}
	_ = db.Upsert("k3", &item3)

	// Query where Src == "1.1.1.1"
	q := bolthold.Where("Src").Eq("1.1.1.1")
	res, err := db.Query(q)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 2)

	// Use QueryMatchFunc to match Dst == "5.5.5.5"
	match := db.QueryMatchFunc("Dst", func(ci *tcp_model.ConnectionItem) (bool, error) {
		return ci.Dst == "5.5.5.5", nil
	})
	res2, err := db.Query(&match)
	assert.NoError(t, err)
	assert.Len(t, res2, 1)
	assert.Equal(t, "5.5.5.5", res2[0].Dst)
}

func TestHashId(t *testing.T) {
	h1 := HashId("abc")
	h2 := HashId("abc")
	h3 := HashId("abcd")
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	// simple sanity: not zero for non-empty
	assert.NotEqual(t, uint32(0), h1)
}
