package nodegraph

import (
	"net/http"

	"github.com/k8spacket/k8spacket/external/db"
	"github.com/k8spacket/k8spacket/external/handlerio"
	"github.com/k8spacket/k8spacket/external/http"
	"github.com/k8spacket/k8spacket/external/k8s"
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/prometheus"
	"github.com/k8spacket/k8spacket/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/modules/nodegraph/stats"
)

func Init(mux *http.ServeMux) modules.IListener[modules.TCPEvent] {

	prometheus.Init()

	handler, _ := db.New[model.ConnectionItem]("tcp_connections")
	repo := &repository.Repository{DbHandler: handler}
	factory := &stats.Factory{}
	service := &Service{repo, factory, &httpclient.HttpClient{}, &k8sclient.K8SClient{}, &handlerio.HandlerIO{}}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	mux.HandleFunc("/nodegraph/connections", controller.ConnectionHandler)
	mux.HandleFunc("/nodegraph/api/health", o11yController.Health)
	mux.HandleFunc("/nodegraph/api/graph/fields", o11yController.NodeGraphFieldsHandler)
	mux.HandleFunc("/nodegraph/api/graph/data", o11yController.NodeGraphDataHandler)

	listener := &Listener{service}

	return listener

}
