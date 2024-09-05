package nodegraph

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/k8spacket/k8spacket/external/db"
	"github.com/k8spacket/k8spacket/external/handlerio"
	"github.com/k8spacket/k8spacket/external/http"
	"github.com/k8spacket/k8spacket/external/k8s"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/modules/nodegraph/stats"
)

type Service struct {
	repo       repository.IRepository[model.ConnectionItem]
	factory    stats.IFactory
	httpClient httpclient.IHttpClient
	k8sClient  k8sclient.IK8SClient
	handlerIO  handlerio.IHandlerIO
}

var connectionItemsMutex = sync.RWMutex{}

func (service *Service) update(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64) {
	connectionItemsMutex.Lock()
	var id = strconv.Itoa(int(db.HashId(fmt.Sprintf("%s-%s", src, dst))))
	var connection = service.repo.Read(id)
	if (model.ConnectionItem{} == connection) {
		connection = *&model.ConnectionItem{Src: src, Dst: dst}
	}
	connection.SrcName = srcName
	connection.SrcNamespace = srcNamespace
	connection.DstName = dstName
	connection.DstNamespace = dstNamespace
	connection.ConnCount++
	if persistent {
		connection.ConnPersistent++
	}
	connection.BytesSent += bytesSent
	connection.BytesReceived += bytesReceived
	connection.Duration += duration
	if duration > connection.MaxDuration {
		connection.MaxDuration = duration
	}
	connection.LastSeen = time.Now()
	service.repo.Set(id, &connection)
	connectionItemsMutex.Unlock()
}

func (service *Service) getConnections(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {

	slog.Info("[api:params]",
		"patternNs", patternNs,
		"patternIn", patternIn,
		"patternEx", patternEx,
		"from", from.Format(time.DateTime),
		"to", to.Format(time.DateTime))

	return service.repo.Query(from, to, patternNs, patternIn, patternEx)
}

func (service *Service) buildO11yResponse(r *http.Request) (model.NodeGraph, error) {
	var k8spacketIps = service.k8sClient.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))

	var in []model.ConnectionItem
	var connectionItems = make(map[string]model.ConnectionItem)

	for _, ip := range k8spacketIps {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/nodegraph/connections?%s", ip, os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), r.URL.Query().Encode()), nil)
		resp, err := service.httpClient.Do(req)

		if err != nil {
			slog.Error("[api] Cannot get stats", "Error", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {

			responseData, err := io.ReadAll(resp.Body)
			if err != nil {
				slog.Error("[api] Cannot read stats response", "Error", err)
				continue
			}

			err = json.Unmarshal(responseData, &in)
			if err != nil {
				slog.Error("[api] Cannot parse stats response", "Error", err)
				continue
			}

			for _, element := range in {
				connectionItems[element.Src+"-"+element.Dst] = element
			}
		}
	}

	var selectedStats = ""
	if len(r.URL.Query()["stats-type"]) > 0 {
		selectedStats = r.URL.Query()["stats-type"][0]
	}
	statsImpl := service.factory.GetStats(selectedStats)

	var connectionEndpoints = make(map[string]model.ConnectionEndpoint)
	prepareConnections(connectionItems, connectionEndpoints)
	return buildApiResponse(connectionItems, connectionEndpoints, statsImpl), nil

}

func (service *Service) getO11yStatsConfig(statsType string) (string, error) {
	jsonFile, err := service.handlerIO.ReadFile("fields.json")
	if err != nil {
		slog.Error("Cannot read file", "Error", err.Error())
		return "", err
	}

	config := service.factory.GetStats(statsType).GetConfig()

	response := string(jsonFile)
	response = strings.ReplaceAll(response, "{{mainStatDisplayName}}", config.MainStat.DisplayName)
	response = strings.ReplaceAll(response, "{{secondaryStatDisplayName}}", config.SecondaryStat.DisplayName)
	response = strings.ReplaceAll(response, "{{arc1color}}", config.Arc1.Color)
	response = strings.ReplaceAll(response, "{{arc1DisplayName}}", config.Arc1.DisplayName)
	response = strings.ReplaceAll(response, "{{arc2color}}", config.Arc2.Color)
	response = strings.ReplaceAll(response, "{{arc2DisplayName}}", config.Arc2.DisplayName)

	return response, nil
}

func prepareConnections(connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint) {

	for _, conn := range connectionItems {
		var connEndpointSrc = connectionEndpoints[conn.Src]
		if (model.ConnectionEndpoint{} == connEndpointSrc) {
			connEndpointSrc = *&model.ConnectionEndpoint{Ip: conn.Src, Name: conn.SrcName, Namespace: conn.SrcNamespace, ConnCount: 0, ConnPersistent: 0, BytesSent: 0, BytesReceived: 0, Duration: 0, MaxDuration: 0}
		}
		connEndpointSrc.BytesSent += conn.BytesSent
		connEndpointSrc.BytesReceived += conn.BytesReceived
		connectionEndpoints[conn.Src] = connEndpointSrc

		var connEndpointDst = connectionEndpoints[conn.Dst]
		if (model.ConnectionEndpoint{} == connEndpointDst) {
			connEndpointDst = *&model.ConnectionEndpoint{Ip: conn.Dst, Name: conn.DstName, Namespace: conn.DstNamespace, ConnCount: 0, ConnPersistent: 0, BytesSent: 0, BytesReceived: 0, Duration: 0, MaxDuration: 0}
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

func buildApiResponse(connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint, statsImpl stats.IStats) model.NodeGraph {

	var nodeArray []model.Node
	var edgeArray []model.Edge

	for _, conn := range connectionItems {
		nodeArray = fillNodesArray(conn.Src, nodeArray, connectionEndpoints, statsImpl)
		nodeArray = fillNodesArray(conn.Dst, nodeArray, connectionEndpoints, statsImpl)
		edgeArray = fillEdgesArray(conn.Src+"-"+conn.Dst, edgeArray, connectionItems, statsImpl)
	}

	return model.NodeGraph{Nodes: nodeArray, Edges: edgeArray}
}

func fillNodesArray(id string, nodeArray []model.Node, connectionEndpoints map[string]model.ConnectionEndpoint, statsImpl stats.IStats) []model.Node {
	var connEndpoint = connectionEndpoints[id]
	var node = model.Node{}
	node.Id = id
	node.Title = connEndpoint.Name
	node.SubTitle = connEndpoint.Ip
	statsImpl.FillNodeStats(&node, connEndpoint)
	nodeArray = append(nodeArray, node)
	return nodeArray
}

func fillEdgesArray(id string, edgeArray []model.Edge, connectionItems map[string]model.ConnectionItem, statsImpl stats.IStats) []model.Edge {
	var connItem = connectionItems[id]
	var edge = model.Edge{}
	edge.Id = id
	edge.Source = connItem.Src
	edge.Target = connItem.Dst
	statsImpl.FillEdgeStats(&edge, connItem)
	edgeArray = append(edgeArray, edge)
	return edgeArray
}
