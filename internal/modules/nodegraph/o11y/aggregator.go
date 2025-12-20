package o11y

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/stats"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
)

func aggregateConnections(ctx context.Context, podIPs []string, query url.Values, port string, client httpclient.Client) []model.ConnectionItem {
	if len(podIPs) == 0 {
		return nil
	}

	const maxConcurrent = 5
	const requestTimeout = 5 * time.Second

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var all []model.ConnectionItem

	for _, ip := range podIPs {
		ip := ip
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("http://%s:%s/nodegraph/connections?%s", ip, port, query.Encode()), nil)
			if err != nil {
				slog.Error("[api] Cannot get stats", "Error", err)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				slog.Error("[api] Cannot get stats", "Error", err)
				return
			}
			//defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				slog.Error("[api] Cannot get stats", "Error", fmt.Errorf("peer %s status %d", ip, resp.StatusCode))
				return
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				slog.Error("[api] Cannot read stats response", "Error", err)
				return
			}

			var fetched []model.ConnectionItem
			if err := json.Unmarshal(data, &fetched); err != nil {
				slog.Error("[api] Cannot parse stats response", "Error", err)
				return
			}

			mu.Lock()
			all = append(all, fetched...)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return all
}

func prepareConnections(connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint) {

	for _, conn := range connectionItems {
		var srcEndpoint = connectionEndpoints[conn.Src]
		if (model.ConnectionEndpoint{} == srcEndpoint) {
			srcEndpoint = model.ConnectionEndpoint{Ip: conn.Src, Name: conn.SrcName, Namespace: conn.SrcNamespace, ConnCount: 0, ConnPersistent: 0, BytesSent: 0, BytesReceived: 0, Duration: 0, MaxDuration: 0}
		}
		srcEndpoint.BytesSent += conn.BytesSent
		srcEndpoint.BytesReceived += conn.BytesReceived
		connectionEndpoints[conn.Src] = srcEndpoint

		var dstEndpoint = connectionEndpoints[conn.Dst]
		if (model.ConnectionEndpoint{} == dstEndpoint) {
			dstEndpoint = model.ConnectionEndpoint{Ip: conn.Dst, Name: conn.DstName, Namespace: conn.DstNamespace, ConnCount: 0, ConnPersistent: 0, BytesSent: 0, BytesReceived: 0, Duration: 0, MaxDuration: 0}
		}
		dstEndpoint.ConnCount += conn.ConnCount
		dstEndpoint.ConnPersistent += conn.ConnPersistent
		dstEndpoint.BytesSent += conn.BytesReceived
		dstEndpoint.BytesReceived += conn.BytesSent
		dstEndpoint.Duration += conn.Duration
		if conn.MaxDuration > dstEndpoint.MaxDuration {
			dstEndpoint.MaxDuration = conn.MaxDuration
		}
		connectionEndpoints[conn.Dst] = dstEndpoint
	}
}

func buildApiResponse(connectionItems map[string]model.ConnectionItem, connectionEndpoints map[string]model.ConnectionEndpoint, statsImpl stats.Stats) model.NodeGraph {

	var nodeArray []model.Node
	var edgeArray []model.Edge

	for _, item := range connectionEndpoints {
		nodeArray = fillNodesArray(item.Ip, nodeArray, connectionEndpoints, statsImpl)
	}

	for _, item := range connectionItems {
		edgeArray = fillEdgesArray(item.Src+"-"+item.Dst, edgeArray, connectionItems, statsImpl)
	}

	return model.NodeGraph{Nodes: nodeArray, Edges: edgeArray}
}

func fillNodesArray(id string, nodeArray []model.Node, connectionEndpoints map[string]model.ConnectionEndpoint, statsImpl stats.Stats) []model.Node {
	var connEndpoint = connectionEndpoints[id]
	var node = model.Node{}
	node.Id = id
	node.Title = connEndpoint.Name
	node.SubTitle = connEndpoint.Ip
	statsImpl.FillNodeStats(&node, connEndpoint)
	nodeArray = append(nodeArray, node)
	return nodeArray
}

func fillEdgesArray(id string, edgeArray []model.Edge, connectionItems map[string]model.ConnectionItem, statsImpl stats.Stats) []model.Edge {
	var connItem = connectionItems[id]
	var edge = model.Edge{}
	edge.Id = id
	edge.Source = connItem.Src
	edge.Target = connItem.Dst
	statsImpl.FillEdgeStats(&edge, connItem)
	edgeArray = append(edgeArray, edge)
	return edgeArray
}
