package nodegraph

import (
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/db"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/prometheus"
	"github.com/k8spacket/k8spacket/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/modules/nodegraph/stats"
	"net/http"
)

func Init() modules.IListener[modules.TCPEvent] {

	prometheus.Init()

	handler, _ := db.New[model.ConnectionItem]("tcp_connections")
	repo := &repository.Repository{DbHandler: handler}
	factory := &stats.Factory{}
	service := &Service{repo, factory}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	http.HandleFunc("/nodegraph/connections", controller.ConnectionHandler)
	http.HandleFunc("/nodegraph/api/health", o11yController.Health)
	http.HandleFunc("/nodegraph/api/graph/fields", o11yController.NodeGraphFieldsHandler)
	http.HandleFunc("/nodegraph/api/graph/data", o11yController.NodeGraphDataHandler)

	listener := &Listener{service}

	return listener

}
