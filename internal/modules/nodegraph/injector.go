package nodegraph

import (
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
	repo := &repository.DbRepository{DbHandler: handler}
	factory := &stats.StatsFactory{}
	service := &NodegraphService{repo: repo, factory: factory, httpClient: &httpclient.HttpClient{}, k8sClient: &k8sclient.K8SClient{}, resource: &resource.FileResource{}}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	mux.HandleFunc("/nodegraph/connections", controller.ConnectionHandler)
	mux.HandleFunc("/nodegraph/api/health", o11yController.Health)
	mux.HandleFunc("/nodegraph/api/graph/fields", o11yController.NodeGraphFieldsHandler)
	mux.HandleFunc("/nodegraph/api/graph/data", o11yController.NodeGraphDataHandler)

	listener := &TcpListener{service}

	return listener

}
