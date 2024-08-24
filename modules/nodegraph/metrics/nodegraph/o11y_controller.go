package nodegraph

import (
	"encoding/json"
	"fmt"
	"github.com/k8spacket/k8s-api/v2"
	nodegraph_log "github.com/k8spacket/k8spacket/modules/nodegraph/log"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph/stats"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

func NodeGraphFieldsHandler(w http.ResponseWriter, r *http.Request) {
	jsonFile, err := os.ReadFile("fields.json")
	if err != nil {
		nodegraph_log.LOGGER.Print(err.Error())
		os.Exit(1)
	}

	var selectedStats = ""
	if len(r.URL.Query()["stats-type"]) > 0 {
		selectedStats = r.URL.Query()["stats-type"][0]
	}

	config := stats.GetConfig(selectedStats)

	response := string(jsonFile)
	response = strings.ReplaceAll(response, "{{mainStatDisplayName}}", config.MainStat.DisplayName)
	response = strings.ReplaceAll(response, "{{secondaryStatDisplayName}}", config.SecondaryStat.DisplayName)
	response = strings.ReplaceAll(response, "{{arc1color}}", config.Arc1.Color)
	response = strings.ReplaceAll(response, "{{arc1DisplayName}}", config.Arc1.DisplayName)
	response = strings.ReplaceAll(response, "{{arc2color}}", config.Arc2.Color)
	response = strings.ReplaceAll(response, "{{arc2DisplayName}}", config.Arc2.DisplayName)

	w.WriteHeader(200)
	w.Write([]byte(response))
}

func NodeGraphDataHandler(w http.ResponseWriter, r *http.Request) {
	var k8spacketIps = k8s.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))

	var in []model.ConnectionItem
	var connectionItems = make(map[string]model.ConnectionItem)

	for _, ip := range k8spacketIps {
		resp, err := http.Get(fmt.Sprintf("http://%s:%s/nodegraph/connections?%s", ip, os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), r.URL.Query().Encode()))

		if err != nil {
			nodegraph_log.LOGGER.Printf("[api] Cannot get stats: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			nodegraph_log.LOGGER.Printf("[api] Cannot read stats response: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = json.Unmarshal(responseData, &in)
		if err != nil {
			nodegraph_log.LOGGER.Printf("[api] Cannot parse stats response: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		for _, element := range in {
			connectionItems[element.Src+"-"+element.Dst] = element
		}
	}

	var connectionEndpoints = make(map[string]model.ConnectionEndpoint)
	prepareConnections(connectionItems, connectionEndpoints)
	buildApiResponse(w, connectionItems, connectionEndpoints, r.URL.Query())

}

func prepareConnections(connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint) {

	for _, conn := range connectionItems {
		var connEndpointSrc = connectionEndpoints[conn.Src]
		if (model.ConnectionEndpoint{} == connEndpointSrc) {
			connEndpointSrc = *&model.ConnectionEndpoint{conn.Src, conn.SrcName, conn.SrcNamespace, 0, 0, 0, 0, 0, 0}
		}
		connEndpointSrc.BytesSent += conn.BytesSent
		connEndpointSrc.BytesReceived += conn.BytesReceived
		connectionEndpoints[conn.Src] = connEndpointSrc

		var connEndpointDst = connectionEndpoints[conn.Dst]
		if (model.ConnectionEndpoint{} == connEndpointDst) {
			connEndpointDst = *&model.ConnectionEndpoint{conn.Dst, conn.DstName, conn.DstNamespace, 0, 0, 0, 0, 0, 0}
		}
		connEndpointDst.ConnCount += conn.ConnCount
		connEndpointDst.ConnPersistent += conn.ConnPersistent
		connEndpointDst.BytesSent += conn.BytesReceived
		connEndpointDst.BytesReceived += conn.BytesSent
		connEndpointDst.Duration += conn.Duration
		if conn.MaxDuration > connEndpointDst.MaxDuration {
			connEndpointDst.MaxDuration = conn.MaxDuration
		}
		connectionEndpoints[conn.Dst] = connEndpointDst
	}
}

func buildApiResponse(w http.ResponseWriter, connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint, query url.Values) {

	var selectedStats = ""
	if len(query["stats-type"]) > 0 {
		selectedStats = query["stats-type"][0]
	}

	var nodeArray []model.Node
	var edgeArray []model.Edge
	for _, conn := range connectionItems {
		nodeArray = fillNodesArray(conn.Src, nodeArray, connectionEndpoints, selectedStats)
		nodeArray = fillNodesArray(conn.Dst, nodeArray, connectionEndpoints, selectedStats)
		edgeArray = fillEdgesArray(conn.Src+"-"+conn.Dst, edgeArray, connectionItems, selectedStats)
	}

	response := model.NodeGraph{nodeArray, edgeArray}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		nodegraph_log.LOGGER.Printf("[api] Cannot prepare stats response: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fillNodesArray(id string, nodeArray []model.Node, connectionEndpoints map[string]model.ConnectionEndpoint, statsType string) []model.Node {
	var connEndpoint = connectionEndpoints[id]
	var node = model.Node{}
	node.Id = id
	node.Title = connEndpoint.Name
	node.SubTitle = connEndpoint.Ip
	stats.FillNodeStats(statsType, &node, connEndpoint)
	nodeArray = append(nodeArray, node)
	return nodeArray
}

func fillEdgesArray(id string, edgeArray []model.Edge, connectionItems map[string]model.ConnectionItem, statsType string) []model.Edge {
	var connItem = connectionItems[id]
	var edge = model.Edge{}
	edge.Id = id
	edge.Source = connItem.Src
	edge.Target = connItem.Dst
	stats.FillEdgeStats(statsType, &edge, connItem)
	edgeArray = append(edgeArray, edge)
	return edgeArray
}
