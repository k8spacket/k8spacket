package bytes

import (
	"fmt"
	"github.com/inhies/go-bytesize"
	"github.com/k8spacket/metrics/nodegraph/model"
)

func GetConfig() model.Config {
	return model.Config{Arc1: model.DisplayConfig{DisplayName: "Bytes received", Color: "blue"},
		Arc2:          model.DisplayConfig{DisplayName: "Bytes responded", Color: "yellow"},
		MainStat:      model.DisplayConfig{DisplayName: "Bytes received"},
		SecondaryStat: model.DisplayConfig{DisplayName: "Bytes responded"}}
}

func FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint) {
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

func FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem) {
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
