package nodegraph

import (
	"net/http"
	"sync"

	"github.com/k8spacket/k8spacket/internal/infra/db"
	"github.com/k8spacket/k8spacket/internal/infra/handlerio"
	httpclient "github.com/k8spacket/k8spacket/internal/infra/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/infra/k8s"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/prometheus"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/repository"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/stats"
	"github.com/k8spacket/k8spacket/pkg/api"
	"github.com/k8spacket/k8spacket/pkg/events"
)

func Init(mux *http.ServeMux) api.IListener[events.TCPEvent] {

	prometheus.Init()

	handler, _ := db.New[model.ConnectionItem]("tcp_connections")
	repo := &repository.Repository{DbHandler: handler}
	factory := &stats.Factory{}
	service := &Service{repo, factory, &httpclient.HttpClient{}, &k8sclient.K8SClient{}, &handlerio.HandlerIO{}, sync.Mutex{}}
	controller := &Controller{service}
	o11yController := &O11yController{service}

	mux.HandleFunc("/nodegraph/connections", controller.ConnectionHandler)
	mux.HandleFunc("/nodegraph/api/health", o11yController.Health)
	mux.HandleFunc("/nodegraph/api/graph/fields", o11yController.NodeGraphFieldsHandler)
	mux.HandleFunc("/nodegraph/api/graph/data", o11yController.NodeGraphDataHandler)

	listener := &Listener{service}

	return listener

}
