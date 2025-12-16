package repository

import (
	"log/slog"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
)

type DbRepository struct {
	DbConnectionHandler db.Db[model.TLSConnection]
	DbDetailsHandler    db.Db[model.TLSDetails]
}

func (repository *DbRepository) Query(from time.Time, to time.Time) []model.TLSConnection {

	query := repository.DbConnectionHandler.QueryMatchFunc("Src", func(record *model.TLSConnection) (bool, error) {
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
		slog.Error("[db:tls_connections:Query]", "Error", err)
		return []model.TLSConnection{}
	}
	return result
}

func (repository *DbRepository) UpsertConnection(key string, value *model.TLSConnection) {
	err := repository.DbConnectionHandler.Upsert(key, value)
	if err != nil {
		slog.Error("[db:tls_connections:Upsert]", "Error", err)
	}
}

func (repository *DbRepository) Read(key string) model.TLSDetails {
	result, err := repository.DbDetailsHandler.Read(key)
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
	err := repository.DbDetailsHandler.Upsert(key, value)
	if err != nil {
		slog.Error("[db:tls_details:Upsert]", "Error", err)
	}
}
