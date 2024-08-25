package repository

import (
	"github.com/k8spacket/k8spacket/modules/db"
	nodegraph_log "github.com/k8spacket/k8spacket/modules/nodegraph/log"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/timshannon/bolthold"
	"regexp"
	"time"
)

type Repository struct {
	DbHandler db.IDBHandler[model.ConnectionItem]
}

func (repository *Repository) Read(key string) model.ConnectionItem {
	result, err := repository.DbHandler.Read(key)
	if err != nil {
		// can happen, silent
		return model.ConnectionItem{}
	}
	return result
}

func (repository *Repository) Query(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {

	query := *bolthold.Where("Src").MatchFunc(func(ra *bolthold.RecordAccess) (bool, error) {
		record := ra.Record().(*model.ConnectionItem)
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

	result, err := repository.DbHandler.Query(&query)
	if err != nil {
		nodegraph_log.LOGGER.Printf("[db:tcp_connections:Query] Error: %+v", err)
		return []model.ConnectionItem{}
	}
	return result
}

func (repository *Repository) Set(key string, value *model.ConnectionItem) {
	err := repository.DbHandler.Upsert(key, *value)
	if err != nil {
		nodegraph_log.LOGGER.Printf("[db:tcp_connections:Upsert] Error: %+v", err)
	}
}
