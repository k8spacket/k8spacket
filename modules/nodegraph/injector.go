package nodegraph

import (
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/db"
	nodegraph_log "github.com/k8spacket/k8spacket/modules/nodegraph/log"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/repository"
	"net/http"
)

func Init() modules.IListener[modules.TCPEvent] {

	nodegraph_log.BuildLogger()

	handler, _ := db.New[model.ConnectionItem]("tcp_connections")
	repo := &repository.Repository{DbHandler: handler}
	service := &Service{repo}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	http.HandleFunc("/nodegraph/connections", controller.ConnectionHandler)
	http.HandleFunc("/nodegraph/api/health", o11yController.Health)
	http.HandleFunc("/nodegraph/api/graph/fields", o11yController.NodeGraphFieldsHandler)
	http.HandleFunc("/nodegraph/api/graph/data", o11yController.NodeGraphDataHandler)

	listener := &Listener{service}

	return listener

}
