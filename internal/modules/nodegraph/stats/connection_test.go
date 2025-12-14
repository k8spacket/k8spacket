package stats

import (
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

func TestConnectionGetConfig(t *testing.T) {
	want := model.Config{Arc1: model.DisplayConfig{DisplayName: "Persistent connections", Color: "green"},
		Arc2:          model.DisplayConfig{DisplayName: "Short-lived connections", Color: "red"},
		MainStat:      model.DisplayConfig{DisplayName: "All connections "},
		SecondaryStat: model.DisplayConfig{DisplayName: "Persistent connections "}}

	connection := &Connection{}

	result := connection.GetConfig()

	assert.EqualValues(t, want, result)

}

func TestConnectionFillNodeStats(t *testing.T) {

	var tests = []struct {
		connectionEndpoint model.ConnectionEndpoint
		want               *model.Node
	}{
		{model.ConnectionEndpoint{ConnCount: 100, ConnPersistent: 10}, &model.Node{MainStat: "all: 100", SecondaryStat: "persistent: 10", Arc1: 0.1, Arc2: 0.9, Arc3: 0}},
		{model.ConnectionEndpoint{}, &model.Node{MainStat: "all: N/A", SecondaryStat: "persistent: N/A"}},
		{model.ConnectionEndpoint{ConnCount: 0}, &model.Node{MainStat: "all: N/A", SecondaryStat: "persistent: N/A"}},
	}

	connection := &Connection{}

	for _, test := range tests {
		t.Run(test.want.Id, func(t *testing.T) {
			t.Parallel()

			node := &model.Node{}
			connection.FillNodeStats(node, test.connectionEndpoint)

			assert.EqualValues(t, test.want, node)
		},
		)
	}
}

func TestConnectionFillEdgeStats(t *testing.T) {
	var tests = []struct {
		ConnectionItem model.ConnectionItem
		want           *model.Edge
	}{
		{model.ConnectionItem{ConnCount: 100, ConnPersistent: 20}, &model.Edge{MainStat: "all: 100", SecondaryStat: "persistent: 20"}},
		{model.ConnectionItem{}, &model.Edge{MainStat: "all: 0", SecondaryStat: "persistent: 0"}},
	}

	connection := &Connection{}

	for _, test := range tests {
		t.Run(test.want.Id, func(t *testing.T) {
			t.Parallel()

			edge := &model.Edge{}
			connection.FillEdgeStats(edge, test.ConnectionItem)

			assert.EqualValues(t, test.want, edge)
		},
		)
	}
}
