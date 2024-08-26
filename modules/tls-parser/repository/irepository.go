package repository

import (
	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"time"
)

type IRepository interface {
	Query(from time.Time, to time.Time) []model.TLSConnection
	UpsertConnection(key string, value *model.TLSConnection)
	Read(key string) model.TLSDetails
	UpsertDetails(key string, value *model.TLSDetails, fn fn)
}
