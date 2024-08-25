package nodegraph

import (
	"encoding/json"
	nodegraph_log "github.com/k8spacket/k8spacket/modules/nodegraph/log"
	"net/http"
)

type O11yController struct {
	service IService
}

func (o11yController *O11yController) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

func (o11yController *O11yController) NodeGraphFieldsHandler(w http.ResponseWriter, r *http.Request) {
	response, err := o11yController.service.GetO11yStatsConfig(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(response))
}

func (o11yController *O11yController) NodeGraphDataHandler(w http.ResponseWriter, r *http.Request) {
	nodegraph, err := o11yController.service.BuildO11yResponse(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err = json.NewEncoder(w).Encode(nodegraph)
	if err != nil {
		nodegraph_log.LOGGER.Printf("[api] Cannot prepare stats response: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
