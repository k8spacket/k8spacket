package stats

import (
	"fmt"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
)

type Connection struct {
	IStats
}

func (connection *Connection) GetConfig() model.Config {
	return model.Config{Arc1: model.DisplayConfig{DisplayName: "Persistent connections", Color: "green"},
		Arc2:          model.DisplayConfig{DisplayName: "Short-lived connections", Color: "red"},
		MainStat:      model.DisplayConfig{DisplayName: "All connections "},
		SecondaryStat: model.DisplayConfig{DisplayName: "Persistent connections "}}
}

func (connection *Connection) FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint) {
	if connEndpoint.ConnCount > 0 {
		node.MainStat = fmt.Sprintf("all: %d", connEndpoint.ConnCount)
		node.SecondaryStat = fmt.Sprintf("persistent: %d", connEndpoint.ConnPersistent)
		node.Arc1 = float64(connEndpoint.ConnPersistent) / float64(connEndpoint.ConnCount)
		node.Arc2 = (float64(connEndpoint.ConnCount) - float64(connEndpoint.ConnPersistent)) / float64(connEndpoint.ConnCount)
	} else {
		node.MainStat = fmt.Sprint("all: N/A")
		node.SecondaryStat = fmt.Sprint("persistent: N/A")
	}
}

func (connection *Connection) FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem) {
	edge.MainStat = fmt.Sprintf("all: %d", connItem.ConnCount)
	edge.SecondaryStat = fmt.Sprintf("persistent: %d", connItem.ConnPersistent)
}
