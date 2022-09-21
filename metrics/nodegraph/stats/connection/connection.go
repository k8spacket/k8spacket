package connection

import (
	"fmt"
	"github.com/k8spacket/metrics/nodegraph/model"
)

func GetConfig() model.Config {
	return model.Config{Arc1: model.DisplayConfig{DisplayName: "Closed connections", Color: "green"},
		Arc2:          model.DisplayConfig{DisplayName: "Unclosed connections", Color: "red"},
		MainStat:      model.DisplayConfig{DisplayName: "All connections "},
		SecondaryStat: model.DisplayConfig{DisplayName: "Closed connections "}}
}

func FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint) {
	if connEndpoint.ConnCount > 0 {
		node.MainStat = fmt.Sprintf("all: %d", connEndpoint.ConnCount)
		node.SecondaryStat = fmt.Sprintf("closed: %d", connEndpoint.ConnClosed)
		node.Arc1 = float64(connEndpoint.ConnClosed) / float64(connEndpoint.ConnCount)
		node.Arc2 = (float64(connEndpoint.ConnCount) - float64(connEndpoint.ConnClosed)) / float64(connEndpoint.ConnCount)
	} else {
		node.MainStat = fmt.Sprint("all: N/A")
		node.SecondaryStat = fmt.Sprint("closed: N/A")
	}
}

func FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem) {
	edge.MainStat = fmt.Sprintf("all: %d", connItem.ConnCount)
	edge.SecondaryStat = fmt.Sprintf("closed: %d", connItem.ConnClosed)
}
