package stats

import (
	"github.com/k8spacket/metrics/nodegraph/model"
	"github.com/k8spacket/metrics/nodegraph/stats/bytes"
	"github.com/k8spacket/metrics/nodegraph/stats/connection"
	"github.com/k8spacket/metrics/nodegraph/stats/duration"
)

func GetConfig(statsType string) model.Config {
	switch statsType {
	case "bytes":
		return bytes.GetConfig()
	case "duration":
		return duration.GetConfig()
	default:
		return connection.GetConfig()
	}
}

func FillNodeStats(statsType string, node *model.Node, connEndpoint model.ConnectionEndpoint) {
	switch statsType {
	case "bytes":
		bytes.FillNodeStats(node, connEndpoint)
	case "duration":
		duration.FillNodeStats(node, connEndpoint)
	default:
		connection.FillNodeStats(node, connEndpoint)
	}
}

func FillEdgeStats(statsType string, edge *model.Edge, connItem model.ConnectionItem) {
	switch statsType {
	case "bytes":
		bytes.FillEdgeStats(edge, connItem)
	case "duration":
		duration.FillEdgeStats(edge, connItem)
	default:
		connection.FillEdgeStats(edge, connItem)
	}
}
