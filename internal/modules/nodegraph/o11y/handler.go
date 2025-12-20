package o11y

import (
	"encoding/json"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/stats"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"github.com/k8spacket/k8spacket/internal/thirdparty/resource"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type O11yHandler struct {
	factory    stats.Factory
	httpClient httpclient.Client
	k8sClient  k8sclient.Client
	resource   resource.Resource
}

func NewO11yHandler(factory stats.Factory, httpClient httpclient.Client, k8sClient k8sclient.Client, resource resource.Resource) *O11yHandler {
	return &O11yHandler{factory: factory, httpClient: httpClient, k8sClient: k8sClient, resource: resource}
}

func (handler *O11yHandler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

func (handler *O11yHandler) NodeGraphFieldsHandler(w http.ResponseWriter, r *http.Request) {
	var selectedStats = ""
	if len(r.URL.Query()["stats-type"]) > 0 {
		selectedStats = r.URL.Query()["stats-type"][0]
	}
	response, err := handler.getO11yStatsConfig(selectedStats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(response))
}

func (handler *O11yHandler) NodeGraphDataHandler(w http.ResponseWriter, r *http.Request) {
	nodegraph, err := handler.buildO11yResponse(r)
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

func (handler *O11yHandler) getO11yStatsConfig(statsType string) (string, error) {
	jsonFile, err := handler.resource.Read("fields.json")
	if err != nil {
		slog.Error("Cannot read file", "Error", err.Error())
		return "", err
	}

	config := handler.factory.GetStats(statsType).GetConfig()

	response := string(jsonFile)
	response = strings.ReplaceAll(response, "{{mainStatDisplayName}}", config.MainStat.DisplayName)
	response = strings.ReplaceAll(response, "{{secondaryStatDisplayName}}", config.SecondaryStat.DisplayName)
	response = strings.ReplaceAll(response, "{{arc1color}}", config.Arc1.Color)
	response = strings.ReplaceAll(response, "{{arc1DisplayName}}", config.Arc1.DisplayName)
	response = strings.ReplaceAll(response, "{{arc2color}}", config.Arc2.Color)
	response = strings.ReplaceAll(response, "{{arc2DisplayName}}", config.Arc2.DisplayName)

	return response, nil
}

func (handler *O11yHandler) buildO11yResponse(r *http.Request) (model.NodeGraph, error) {
	var k8spacketIps = handler.k8sClient.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))
	var connectionItems = make(map[string]model.ConnectionItem)

	fetched := aggregateConnections(r.Context(), k8spacketIps, r.URL.Query(), os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), handler.httpClient)
	for _, element := range fetched {
		connectionItems[element.Src+"-"+element.Dst] = element
	}

	var selectedStats = ""
	if len(r.URL.Query()["stats-type"]) > 0 {
		selectedStats = r.URL.Query()["stats-type"][0]
	}
	statsImpl := handler.factory.GetStats(selectedStats)

	var connectionEndpoints = make(map[string]model.ConnectionEndpoint)
	prepareConnections(connectionItems, connectionEndpoints)
	return buildApiResponse(connectionItems, connectionEndpoints, statsImpl), nil

}
