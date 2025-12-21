package nodegraph

import (
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/backend"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/listener"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/o11y"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/updater"
	"net/http"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/prometheus"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/stats"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
	"github.com/k8spacket/k8spacket/internal/thirdparty/http"
	"github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/k8spacket/k8spacket/internal/thirdparty/resource"
)

func Init(mux *http.ServeMux) modules.Listener[modules.TCPEvent] {

	prometheus.Init()

	handler, _ := db.New[model.ConnectionItem]("tcp_connections")
	repo := repository.NewDbRepository(handler)
	controller := backend.NewHandler(repo)
	o11yController := o11y.NewO11yHandler(&stats.StatsFactory{}, &httpclient.HttpClient{}, &k8sclient.K8SClient{}, &resource.FileResource{})

	mux.HandleFunc("/nodegraph/connections", controller.ConnectionHandler)
	mux.HandleFunc("/nodegraph/api/health", o11yController.Health)
	mux.HandleFunc("/nodegraph/api/graph/fields", o11yController.NodeGraphFieldsHandler)
	mux.HandleFunc("/nodegraph/api/graph/data", o11yController.NodeGraphDataHandler)

	nodegraphUpdater := updater.NewUpdater(repo)
	tcpListener := listener.NewListener(nodegraphUpdater)

	return tcpListener

}
