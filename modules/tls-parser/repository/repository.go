package repository

import (
	"github.com/k8spacket/k8spacket/modules/db"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"github.com/timshannon/bolthold"
	"time"
)

type Repository struct {
	DbConnectionHandler db.IDBHandler[model.TLSConnection]
	DbDetailsHandler    db.IDBHandler[model.TLSDetails]
}

func (repository *Repository) Query(from time.Time, to time.Time) []model.TLSConnection {

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

	result, err := repository.DbConnectionHandler.Query(&query)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_connections:Query] Error: %+v", err)
		return []model.TLSConnection{}
	}
	return result
}

func (repository *Repository) UpsertConnection(key string, value *model.TLSConnection) {
	err := repository.DbConnectionHandler.Upsert(key, *value)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_connections:Upsert] Error: %+v", err)
	}
}

func (repository *Repository) Read(key string) model.TLSDetails {
	result, err := repository.DbDetailsHandler.Read(key)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_details:Read] Warn: %+v", err)
		//can happen, silent
		return model.TLSDetails{}
	}
	return result
}

type fn func(newValue *model.TLSDetails, oldValue *model.TLSDetails)

func (repository *Repository) UpsertDetails(key string, value *model.TLSDetails, fn fn) {
	old := repository.Read(key)
	fn(value, &old)
	err := repository.DbDetailsHandler.Upsert(key, *value)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_details:Upsert] Error: %+v", err)
	}
}
