package storer

import (
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
)

type Storer interface {
	StoreInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails)
}
