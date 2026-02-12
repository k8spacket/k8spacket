package stats

import (
	"fmt"
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"time"
)

type Duration struct {
	IStats
}

func (duration *Duration) GetConfig() model.Config {
	return model.Config{Arc1: model.DisplayConfig{DisplayName: "Average duration", Color: "purple"},
		Arc2:          model.DisplayConfig{DisplayName: "Max duration", Color: "white"},
		MainStat:      model.DisplayConfig{DisplayName: "Average duration "},
		SecondaryStat: model.DisplayConfig{DisplayName: "Max duration "}}
}

func (duration *Duration) FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint) {
	if connEndpoint.Duration > 0 || connEndpoint.MaxDuration > 0 {
		var cd = connEndpoint.Duration / float64(connEndpoint.ConnCount)
		if cd >= 0.001 {
			node.MainStat = fmt.Sprintf("avg: %s", time.Duration(cd*float64(time.Millisecond)))
		} else {
			node.MainStat = fmt.Sprint("avg: <0.001s")
		}
		if connEndpoint.MaxDuration >= 0.001 {
			node.SecondaryStat = fmt.Sprintf("max: %s", time.Duration(connEndpoint.MaxDuration*float64(time.Millisecond)))
		} else {
			node.SecondaryStat = fmt.Sprint("max: <0.001s")
		}
		node.Arc1 = cd / connEndpoint.MaxDuration
		node.Arc2 = (connEndpoint.MaxDuration - cd) / connEndpoint.MaxDuration
	} else {
		node.MainStat = fmt.Sprint("avg: N/A")
		node.SecondaryStat = fmt.Sprint("max: N/A")
	}
	node.Arc1Color = "purple"
	node.Arc2Color = "white"
}

func (duration *Duration) FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem) {
	if connItem.Duration > 0 || connItem.MaxDuration > 0 {
		var cd = connItem.Duration / float64(connItem.ConnCount)
		if cd >= 0.001 {
			edge.MainStat = fmt.Sprintf("avg: %s", time.Duration(cd*float64(time.Millisecond)))
		} else {
			edge.MainStat = fmt.Sprint("avg: <0.001s")
		}
		if connItem.MaxDuration >= 0.001 {
			edge.SecondaryStat = fmt.Sprintf("max: %s", time.Duration(connItem.MaxDuration*float64(time.Millisecond)))
		} else {
			edge.SecondaryStat = fmt.Sprint("max: <0.001s")
		}
	} else {
		edge.MainStat = fmt.Sprint("avg: N/A")
		edge.SecondaryStat = fmt.Sprint("max: N/A")
	}
}
