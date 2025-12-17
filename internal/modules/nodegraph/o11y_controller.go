package nodegraph

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type O11yController struct {
	service Service
}

func (o11yController *O11yController) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

func (o11yController *O11yController) NodeGraphFieldsHandler(w http.ResponseWriter, r *http.Request) {
	var selectedStats = ""
	if len(r.URL.Query()["stats-type"]) > 0 {
		selectedStats = r.URL.Query()["stats-type"][0]
	}
	response, err := o11yController.service.getO11yStatsConfig(selectedStats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(response))
}

func (o11yController *O11yController) NodeGraphDataHandler(w http.ResponseWriter, r *http.Request) {
	nodegraph, err := o11yController.service.buildO11yResponse(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err = json.NewEncoder(w).Encode(nodegraph)
	if err != nil {
		slog.Error("[api] Cannot prepare stats response", "Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
