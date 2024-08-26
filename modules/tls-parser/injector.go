package tlsparser

import (
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/db"
	"github.com/k8spacket/k8spacket/modules/tls-parser/certificate"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"github.com/k8spacket/k8spacket/modules/tls-parser/prometheus"
	"github.com/k8spacket/k8spacket/modules/tls-parser/repository"
	"net/http"
)

func Init() modules.IListener[modules.TLSEvent] {

	tls_parser_log.BuildLogger()
	prometheus.Init()

	handlerConnections, _ := db.New[model.TLSConnection]("tls_connections")
	handlerDetails, _ := db.New[model.TLSDetails]("tls_details")
	repo := &repository.Repository{DbConnectionHandler: handlerConnections, DbDetailsHandler: handlerDetails}
	cert := &certificate.Certificate{}
	service := &Service{repo, cert}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	http.HandleFunc("/tlsparser/connections/", controller.TLSConnectionHandler)
	http.HandleFunc("/tlsparser/api/data", o11yController.TLSParserConnectionsHandler)
	http.HandleFunc("/tlsparser/api/data/", o11yController.TLSParserConnectionDetailsHandler)

	listener := &Listener{service}

	return listener

}
