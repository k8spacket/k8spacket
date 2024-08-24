package tls_connection_db

import (
	"github.com/k8spacket/k8spacket/modules/idb"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
	"github.com/timshannon/bolthold"
	"time"
)

var db, _ = idb.StartDB[model.TLSConnection]("tls_connections")

func Query(from time.Time, to time.Time) []model.TLSConnection {

	query := *bolthold.Where("Src").MatchFunc(func(ra *bolthold.RecordAccess) (bool, error) {
		record := ra.Record().(*model.TLSConnection)
		valid := true
		if !from.IsZero() {
			valid = record.LastSeen.After(from) &&
				valid
		}
		if !to.IsZero() {
			valid = record.LastSeen.Before(to) &&
				valid
		}

		return valid, nil
	})

	result, err := db.Query(&query)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_connections:Query] Error: %+v", err)
		return []model.TLSConnection{}
	}
	return result
}

func Upsert(key string, value *model.TLSConnection) {
	err := db.Upsert(key, *value)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_connections:Upsert] Error: %+v", err)
	}
}
