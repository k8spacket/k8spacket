package tlsparser

import (
	"net/http"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/prometheus"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/update"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	"github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/k8spacket/k8spacket/internal/thirdparty/network"
)

func Init(mux *http.ServeMux) modules.Listener[modules.TLSEvent] {

	prometheus.Init()

	handlerConnections, _ := db.New[model.TLSConnection]("tls_connections")
	handlerDetails, _ := db.New[model.TLSDetails]("tls_details")
	repo := &repository.DbRepository{DbConnectionHandler: handlerConnections, DbDetailsHandler: handlerDetails}
	cert := &update.CertificateUpdater{Network: &network.HttpConnectionInspector{}}
	service := &TlsParserService{repo: repo, updater: cert, httpClient: &httpclient.HttpClient{}, k8sClient: &k8sclient.K8SClient{}}
	controller := &Controller{service: service}
	o11yController := &O11yController{service: service}

	mux.HandleFunc("/tlsparser/connections/", controller.TLSConnectionHandler)
	mux.HandleFunc("/tlsparser/api/data", o11yController.TLSParserConnectionsHandler)
	mux.HandleFunc("/tlsparser/api/data/", o11yController.TLSParserConnectionDetailsHandler)

	listener := &TlsListener{service: service}

	return listener

}
