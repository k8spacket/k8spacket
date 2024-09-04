package tlsparser

import (
	"net/http"

	"github.com/k8spacket/k8spacket/external/db"
	httpclient "github.com/k8spacket/k8spacket/external/http"
	"github.com/k8spacket/k8spacket/external/k8s"
	"github.com/k8spacket/k8spacket/external/network"
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/tls-parser/certificate"
	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"github.com/k8spacket/k8spacket/modules/tls-parser/prometheus"
	"github.com/k8spacket/k8spacket/modules/tls-parser/repository"
)

func Init(mux *http.ServeMux) modules.IListener[modules.TLSEvent] {

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
