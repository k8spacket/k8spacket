package tls_detail_db

import (
	"github.com/k8spacket/k8spacket/modules/idb"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
)

var db, _ = idb.StartDB[model.TLSDetails]("tls_details")

func Read(key string) model.TLSDetails {
	result, err := db.Read(key)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_details:Read] Warn: %+v", err)
		//can happen, silent
		return model.TLSDetails{}
	}
	return result
}

type fn func(newValue *model.TLSDetails, oldValue *model.TLSDetails)

func Upsert(key string, value *model.TLSDetails, fn fn) {
	old := Read(key)
	fn(value, &old)
	err := db.Upsert(key, *value)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[db:tls_details:Upsert] Error: %+v", err)
	}
}
