package tlsparser

import (
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/model"
	"net/url"
)

type IService interface {
	storeInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails)

	getConnection(id string) model.TLSDetails

	filterConnections(query url.Values) []model.TLSConnection

	buildConnectionsResponse(url string) ([]model.TLSConnection, error)

	buildDetailsResponse(url string) (model.TLSDetails, error)
}
