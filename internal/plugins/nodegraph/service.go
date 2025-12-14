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

	"github.com/k8spacket/k8spacket/internal/infra/db"
	"github.com/k8spacket/k8spacket/internal/infra/handlerio"
	httpclient "github.com/k8spacket/k8spacket/internal/infra/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/infra/k8s"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/repository"
	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/stats"
)

type Service struct {
	repo       repository.IRepository[model.ConnectionItem]
	factory    stats.IFactory
	httpClient httpclient.IHttpClient
	k8sClient  k8sclient.IK8SClient
	handlerIO  handlerio.IHandlerIO
	mu         sync.Mutex
}

type ConnectionUpdate struct {
	Src           string
	SrcName       string
	SrcNamespace  string
	Dst           string
	DstName       string
	DstNamespace  string
	Persistent    bool
	BytesSent     float64
	BytesReceived float64
	Duration      float64
	Closed        bool
}

func (service *Service) recordConnection(update ConnectionUpdate) {
	id := service.connectionKey(update.Src, update.Dst)

	service.mu.Lock()
	defer service.mu.Unlock()

	connection := service.repo.Read(id)
	if (model.ConnectionItem{} == connection) {
		connection = model.ConnectionItem{Src: update.Src, Dst: update.Dst}
	}

	connection.SrcName = update.SrcName
	connection.SrcNamespace = update.SrcNamespace
	connection.DstName = update.DstName
	connection.DstNamespace = update.DstNamespace

	if update.Closed {
		connection.ConnCount++
		if update.Persistent {
			connection.ConnPersistent++
		}
		connection.BytesSent += update.BytesSent
		connection.BytesReceived += update.BytesReceived
		connection.Duration += update.Duration
		if update.Duration > connection.MaxDuration {
			connection.MaxDuration = update.Duration
		}
	}

	connection.LastSeen = time.Now()
	service.repo.Set(id, &connection)
}

func (service *Service) connectionKey(src, dst string) string {
	return strconv.Itoa(int(db.HashId(fmt.Sprintf("%s-%s", src, dst))))
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
	connectionItems, err := service.collectConnections(r)
	if err != nil {
		slog.Error("[api] Cannot collect stats", "Error", err)
	}

	selectedStats := ""
	if len(r.URL.Query()["stats-type"]) > 0 {
		selectedStats = r.URL.Query()["stats-type"][0]
	}
	statsImpl := service.factory.GetStats(selectedStats)

	connectionEndpoints := aggregateEndpoints(connectionItems)
	return buildApiResponse(connectionItems, connectionEndpoints, statsImpl), nil

}

func (service *Service) getO11yStatsConfig(statsType string) (string, error) {
	jsonFile, err := service.handlerIO.ReadFile("fields.json")
	if err != nil {
		slog.Error("Cannot read file", "Error", err.Error())
		return "", err
	}

	config := service.factory.GetStats(statsType).GetConfig()

	replacer := strings.NewReplacer(
		"{{mainStatDisplayName}}", config.MainStat.DisplayName,
		"{{secondaryStatDisplayName}}", config.SecondaryStat.DisplayName,
		"{{arc1color}}", config.Arc1.Color,
		"{{arc1DisplayName}}", config.Arc1.DisplayName,
		"{{arc2color}}", config.Arc2.Color,
		"{{arc2DisplayName}}", config.Arc2.DisplayName,
	)

	return replacer.Replace(string(jsonFile)), nil
}

func (service *Service) collectConnections(r *http.Request) (map[string]model.ConnectionItem, error) {
	k8spacketIps := service.k8sClient.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))
	connectionItems := make(map[string]model.ConnectionItem)
	var firstErr error

	for _, ip := range k8spacketIps {
		items, err := service.fetchRemoteConnections(ip, r.URL.RawQuery)
		if err != nil {
			slog.Error("[api] Cannot get stats", "Error", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, element := range items {
			connectionItems[element.Src+"-"+element.Dst] = element
		}
	}

	return connectionItems, firstErr
}

func (service *Service) fetchRemoteConnections(ip string, rawQuery string) ([]model.ConnectionItem, error) {
	url := fmt.Sprintf("http://%s:%s/nodegraph/connections?%s", ip, os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), rawQuery)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := service.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []model.ConnectionItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func aggregateEndpoints(connectionItems map[string]model.ConnectionItem) map[string]model.ConnectionEndpoint {
	connectionEndpoints := make(map[string]model.ConnectionEndpoint)

	for _, conn := range connectionItems {
		src := connectionEndpoints[conn.Src]
		if src == (model.ConnectionEndpoint{}) {
			src = model.ConnectionEndpoint{Ip: conn.Src, Name: conn.SrcName, Namespace: conn.SrcNamespace}
		}
		if src.Name == "" {
			src.Name = conn.SrcName
		}
		if src.Namespace == "" {
			src.Namespace = conn.SrcNamespace
		}
		src.BytesSent += conn.BytesSent
		src.BytesReceived += conn.BytesReceived
		connectionEndpoints[conn.Src] = src

		dst := connectionEndpoints[conn.Dst]
		if dst == (model.ConnectionEndpoint{}) {
			dst = model.ConnectionEndpoint{Ip: conn.Dst, Name: conn.DstName, Namespace: conn.DstNamespace}
		}
		if dst.Name == "" {
			dst.Name = conn.DstName
		}
		if dst.Namespace == "" {
			dst.Namespace = conn.DstNamespace
		}
		dst.ConnCount += conn.ConnCount
		dst.ConnPersistent += conn.ConnPersistent
		dst.BytesSent += conn.BytesReceived
		dst.BytesReceived += conn.BytesSent
		dst.Duration += conn.Duration
		if conn.MaxDuration > dst.MaxDuration {
			dst.MaxDuration = conn.MaxDuration
		}
		connectionEndpoints[conn.Dst] = dst
	}

	return connectionEndpoints
}

func buildApiResponse(connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint, statsImpl stats.IStats) model.NodeGraph {
	nodes := make([]model.Node, 0, len(connectionEndpoints))
	for id, endpoint := range connectionEndpoints {
		node := model.Node{Id: id, Title: endpoint.Name, SubTitle: endpoint.Ip}
		statsImpl.FillNodeStats(&node, endpoint)
		nodes = append(nodes, node)
	}

	edges := make([]model.Edge, 0, len(connectionItems))
	for id, item := range connectionItems {
		edge := model.Edge{Id: id, Source: item.Src, Target: item.Dst}
		statsImpl.FillEdgeStats(&edge, item)
		edges = append(edges, edge)
	}

	return model.NodeGraph{Nodes: nodes, Edges: edges}
}
