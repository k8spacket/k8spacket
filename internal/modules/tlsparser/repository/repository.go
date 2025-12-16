package repository

import (
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"time"
)

type Repository interface {
	Query(from time.Time, to time.Time) []model.TLSConnection
	UpsertConnection(key string, value *model.TLSConnection)
	Read(key string) model.TLSDetails
	UpsertDetails(key string, value *model.TLSDetails, fn Fn)
}
