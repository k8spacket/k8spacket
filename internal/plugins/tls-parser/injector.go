package tlsparser

import (
	"net/http"

	"github.com/k8spacket/k8spacket/internal/infra/db"
	httpclient "github.com/k8spacket/k8spacket/internal/infra/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/infra/k8s"
	"github.com/k8spacket/k8spacket/internal/infra/network"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/certificate"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/model"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/prometheus"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/repository"
	"github.com/k8spacket/k8spacket/pkg/api"
	"github.com/k8spacket/k8spacket/pkg/events"
)

func Init(mux *http.ServeMux) api.IListener[events.TLSEvent] {

	prometheus.Init()

	handlerConnections, _ := db.New[model.TLSConnection]("tls_connections")
	handlerDetails, _ := db.New[model.TLSDetails]("tls_details")
	repo := &repository.Repository{DbConnectionHandler: handlerConnections, DbDetailsHandler: handlerDetails}
	cert := &certificate.Certificate{Network: &network.Network{}}
	service := &Service{repo, cert, &httpclient.HttpClient{}, &k8sclient.K8SClient{}}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	mux.HandleFunc("/tlsparser/connections/", controller.TLSConnectionHandler)
	mux.HandleFunc("/tlsparser/api/data", o11yController.TLSParserConnectionsHandler)
	mux.HandleFunc("/tlsparser/api/data/", o11yController.TLSParserConnectionDetailsHandler)

	listener := &Listener{service}

	return listener

}
