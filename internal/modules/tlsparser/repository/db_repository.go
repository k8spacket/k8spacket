package repository

import (
	"log/slog"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
)

type DbRepository struct {
	dbConnectionHandler db.Db[model.TLSConnection]
	dbDetailsHandler    db.Db[model.TLSDetails]
}

func NewDbRepository(db db.Db[model.TLSConnection], dbDetails db.Db[model.TLSDetails]) *DbRepository {
	return &DbRepository{dbConnectionHandler: db, dbDetailsHandler: dbDetails}
}

func (repository *DbRepository) Query(from time.Time, to time.Time) []model.TLSConnection {

	query := repository.dbConnectionHandler.QueryMatchFunc("Src", func(record *model.TLSConnection) (bool, error) {
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

	result, err := repository.dbConnectionHandler.Query(&query)
	if err != nil {
		slog.Error("[db:tls_connections:Query]", "Error", err)
		return []model.TLSConnection{}
	}
	return result
}

func (repository *DbRepository) UpsertConnection(key string, value *model.TLSConnection) {
	err := repository.dbConnectionHandler.Upsert(key, value)
	if err != nil {
		slog.Error("[db:tls_connections:Upsert]", "Error", err)
	}
}

func (repository *DbRepository) Read(key string) model.TLSDetails {
	result, err := repository.dbDetailsHandler.Read(key)
	if err != nil {
		slog.Warn("[db:tls_details:Read]", "Error", err)
		//can happen, silent
		return model.TLSDetails{}
	}
	return result
}

type Fn func(newValue *model.TLSDetails, oldValue *model.TLSDetails)

func (repository *DbRepository) UpsertDetails(key string, value *model.TLSDetails, fn Fn) {
	old := repository.Read(key)
	fn(value, &old)
	err := repository.dbDetailsHandler.Upsert(key, value)
	if err != nil {
		slog.Error("[db:tls_details:Upsert]", "Error", err)
	}
}
