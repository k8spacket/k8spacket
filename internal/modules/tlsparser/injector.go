package tlsparser

import (
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/backend"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/listener"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/o11y"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/storer"
	"net/http"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/prometheus"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/update"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/k8spacket/k8spacket/internal/thirdparty/network"
)

func Init(mux *http.ServeMux) modules.Listener[modules.TLSEvent] {

	prometheus.Init()

	handlerConnections, _ := db.New[model.TLSConnection]("tls_connections")
	handlerDetails, _ := db.New[model.TLSDetails]("tls_details")
	repo := repository.NewDbRepository(handlerConnections, handlerDetails)
	cert := update.NewUpdater(&network.HttpConnectionInspector{})
	handler := backend.NewHandler(repo)
	o11yHandler := o11y.NewO11yHandler(&httpclient.HttpClient{}, &k8sclient.K8SClient{})

	mux.HandleFunc("/tlsparser/connections/", handler.TLSConnectionHandler)
	mux.HandleFunc("/tlsparser/api/data", o11yHandler.TLSParserConnectionsHandler)
	mux.HandleFunc("/tlsparser/api/data/", o11yHandler.TLSParserConnectionDetailsHandler)

	repositoryStorer := storer.NewStorer(repo, cert)
	tlsListener := listener.NewListener(repositoryStorer)

	return tlsListener

}
