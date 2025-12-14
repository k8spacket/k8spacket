package stats

import (
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

func TestBytesGetConfig(t *testing.T) {
	want := model.Config{Arc1: model.DisplayConfig{DisplayName: "Bytes received", Color: "blue"},
		Arc2:          model.DisplayConfig{DisplayName: "Bytes responded", Color: "yellow"},
		MainStat:      model.DisplayConfig{DisplayName: "Bytes received "},
		SecondaryStat: model.DisplayConfig{DisplayName: "Bytes responded "}}

	bytes := &Bytes{}

	result := bytes.GetConfig()

	assert.EqualValues(t, want, result)

}

func TestBytesFillNodeStats(t *testing.T) {

	var tests = []struct {
		connectionEndpoint model.ConnectionEndpoint
		want               *model.Node
	}{
		{model.ConnectionEndpoint{BytesSent: 300, BytesReceived: 100, Duration: 0.5}, &model.Node{MainStat: "recv: 200.00B/s", SecondaryStat: "resp: 600.00B/s", Arc1: 0.25, Arc2: 0.75, Arc3: 0}},
		{model.ConnectionEndpoint{BytesReceived: 100, Duration: 0.5}, &model.Node{MainStat: "recv: N/A", SecondaryStat: "resp: N/A"}},
		{model.ConnectionEndpoint{BytesSent: 300, Duration: 0.5}, &model.Node{MainStat: "recv: N/A", SecondaryStat: "resp: N/A"}},
		{model.ConnectionEndpoint{BytesSent: 300, BytesReceived: 100}, &model.Node{MainStat: "recv: N/A", SecondaryStat: "resp: N/A"}},
	}

	bytes := &Bytes{}

	for _, test := range tests {
		t.Run(test.want.Id, func(t *testing.T) {
			t.Parallel()

			node := &model.Node{}
			bytes.FillNodeStats(node, test.connectionEndpoint)

			assert.EqualValues(t, test.want, node)
		},
		)
	}
}

func TestBytesFillEdgeStats(t *testing.T) {
	var tests = []struct {
		ConnectionItem model.ConnectionItem
		want           *model.Edge
	}{
		{model.ConnectionItem{BytesSent: 300, BytesReceived: 100, Duration: 0.5}, &model.Edge{MainStat: "sent: 600.00B/s", SecondaryStat: "recv: 200.00B/s"}},
		{model.ConnectionItem{BytesReceived: 100, Duration: 0.5}, &model.Edge{MainStat: "sent: N/A", SecondaryStat: "recv: N/A"}},
		{model.ConnectionItem{BytesSent: 300, Duration: 0.5}, &model.Edge{MainStat: "sent: N/A", SecondaryStat: "recv: N/A"}},
		{model.ConnectionItem{BytesSent: 300, BytesReceived: 100}, &model.Edge{MainStat: "sent: N/A", SecondaryStat: "recv: N/A"}},
	}

	bytes := &Bytes{}

	for _, test := range tests {
		t.Run(test.want.Id, func(t *testing.T) {
			t.Parallel()

			edge := &model.Edge{}
			bytes.FillEdgeStats(edge, test.ConnectionItem)

			assert.EqualValues(t, test.want, edge)
		},
		)
	}
}
