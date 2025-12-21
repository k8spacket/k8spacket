package repository

import (
	"log/slog"
	"regexp"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
)

type DbRepository struct {
	dbHandler db.Db[model.ConnectionItem]
}

func NewDbRepository(db db.Db[model.ConnectionItem]) *DbRepository {
	return &DbRepository{dbHandler: db}
}

func (repository *DbRepository) Read(key string) model.ConnectionItem {
	result, err := repository.dbHandler.Read(key)
	if err != nil {
		// can happen, silent
		return model.ConnectionItem{}
	}
	return result
}

func (repository *DbRepository) Query(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {

	query := repository.dbHandler.QueryMatchFunc("Src", func(record *model.ConnectionItem) (bool, error) {
		valid := true
		if !from.IsZero() {
			valid = record.LastSeen.After(from) &&
				valid
		}
		if !to.IsZero() {
			valid = record.LastSeen.Before(to) &&
				valid
		}
		if "" != patternNs.String() {
			valid = (patternNs.Match([]byte(record.SrcNamespace)) ||
				patternNs.Match([]byte(record.DstNamespace))) &&
				valid
		}
		if "" != patternIn.String() {
			valid = (patternIn.Match([]byte(record.Src)) ||
				patternIn.Match([]byte(record.SrcName)) ||
				patternIn.Match([]byte(record.Dst)) ||
				patternIn.Match([]byte(record.DstName))) &&
				valid
		}
		if "" != patternEx.String() {
			valid = !(patternEx.Match([]byte(record.Src)) ||
				patternEx.Match([]byte(record.SrcName)) ||
				patternEx.Match([]byte(record.Dst)) ||
				patternEx.Match([]byte(record.DstName))) &&
				valid
		}

		return valid, nil
	})

	result, err := repository.dbHandler.Query(&query)
	if err != nil {
		slog.Error("[db:tcp_connections:Query]", "Error", err)
		return []model.ConnectionItem{}
	}
	return result
}

func (repository *DbRepository) Set(key string, value *model.ConnectionItem) {
	err := repository.dbHandler.Upsert(key, value)
	if err != nil {
		slog.Error("[db:tcp_connections:Upsert]", "Error", err)
	}
}
