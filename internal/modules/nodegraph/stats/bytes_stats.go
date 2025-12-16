package stats

import (
	"fmt"
	"github.com/inhies/go-bytesize"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
)

type BytesStats struct {
	Stats
}

func (bytes *BytesStats) GetConfig() model.Config {
	return model.Config{Arc1: model.DisplayConfig{DisplayName: "BytesStats received", Color: "blue"},
		Arc2:          model.DisplayConfig{DisplayName: "BytesStats responded", Color: "yellow"},
		MainStat:      model.DisplayConfig{DisplayName: "BytesStats received "},
		SecondaryStat: model.DisplayConfig{DisplayName: "BytesStats responded "}}
}

func (bytes *BytesStats) FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint) {
	if connEndpoint.BytesSent > 0 && connEndpoint.BytesReceived > 0 && connEndpoint.Duration > 0 {
		var sps = bytesize.New(connEndpoint.BytesSent / connEndpoint.Duration)
		var rps = bytesize.New(connEndpoint.BytesReceived / connEndpoint.Duration)
		node.MainStat = fmt.Sprintf("recv: %s/s", rps)
		node.SecondaryStat = fmt.Sprintf("resp: %s/s", sps)
		node.Arc1 = connEndpoint.BytesReceived / (connEndpoint.BytesSent + connEndpoint.BytesReceived)
		node.Arc2 = connEndpoint.BytesSent / (connEndpoint.BytesSent + connEndpoint.BytesReceived)
	} else {
		node.MainStat = fmt.Sprint("recv: N/A")
		node.SecondaryStat = fmt.Sprint("resp: N/A")
	}
}

func (bytes *BytesStats) FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem) {
	if connItem.BytesSent > 0 && connItem.BytesReceived > 0 && connItem.Duration > 0 {
		var sps = bytesize.New(connItem.BytesSent / connItem.Duration)
		var rps = bytesize.New(connItem.BytesReceived / connItem.Duration)
		edge.MainStat = fmt.Sprintf("sent: %s/s", sps)
		edge.SecondaryStat = fmt.Sprintf("recv: %s/s", rps)
	} else {
		edge.MainStat = fmt.Sprint("sent: N/A")
		edge.SecondaryStat = fmt.Sprint("recv: N/A")
	}
}
