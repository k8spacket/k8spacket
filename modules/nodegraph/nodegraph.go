package nodegraph

import (
	"fmt"
	"net/http"

	nodegraph_log "github.com/k8spacket/k8spacket/modules/nodegraph/log"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph"
)

func init() {
	fmt.Println("elo")
	
	nodegraph_log.BuildLogger()

	http.HandleFunc("/nodegraph/connections", nodegraph.ConnectionHandler)
	http.HandleFunc("/nodegraph/api/health", nodegraph.Health)
	http.HandleFunc("/nodegraph/api/graph/fields", nodegraph.NodeGraphFieldsHandler)
	http.HandleFunc("/nodegraph/api/graph/data", nodegraph.NodeGraphDataHandler)
}
