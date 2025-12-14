package stats

import (
	"testing"

	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/model"
	"github.com/stretchr/testify/assert"
)

func TestDurationGetConfig(t *testing.T) {
	want := model.Config{Arc1: model.DisplayConfig{DisplayName: "Average duration", Color: "purple"},
		Arc2:          model.DisplayConfig{DisplayName: "Max duration", Color: "white"},
		MainStat:      model.DisplayConfig{DisplayName: "Average duration "},
		SecondaryStat: model.DisplayConfig{DisplayName: "Max duration "}}

	duration := &Duration{}

	result := duration.GetConfig()

	assert.EqualValues(t, want, result)

}

func TestDurationFillNodeStats(t *testing.T) {

	var tests = []struct {
		connectionEndpoint model.ConnectionEndpoint
		want               *model.Node
	}{
		{model.ConnectionEndpoint{Duration: 100, MaxDuration: 200, ConnCount: 2}, &model.Node{MainStat: "avg: 50ms", SecondaryStat: "max: 200ms", Arc1: 0.25, Arc2: 0.75, Arc3: 0}},
		{model.ConnectionEndpoint{Duration: 0.0001, MaxDuration: 0.0008, ConnCount: 1}, &model.Node{MainStat: "avg: <0.001s", SecondaryStat: "max: <0.001s", Arc1: 0.125, Arc2: 0.875, Arc3: 0}},
		{model.ConnectionEndpoint{ConnCount: 1}, &model.Node{MainStat: "avg: N/A", SecondaryStat: "max: N/A"}},
	}

	duration := &Duration{}

	for _, test := range tests {
		t.Run(test.want.Id, func(t *testing.T) {
			t.Parallel()

			node := &model.Node{}
			duration.FillNodeStats(node, test.connectionEndpoint)

			assert.EqualValues(t, test.want, node)
		},
		)
	}
}

func TestDurationFillEdgeStats(t *testing.T) {
	var tests = []struct {
		ConnectionItem model.ConnectionItem
		want           *model.Edge
	}{
		{model.ConnectionItem{Duration: 100, MaxDuration: 200, ConnCount: 2}, &model.Edge{MainStat: "avg: 50ms", SecondaryStat: "max: 200ms"}},
		{model.ConnectionItem{Duration: 0.0001, MaxDuration: 0.0008, ConnCount: 1}, &model.Edge{MainStat: "avg: <0.001s", SecondaryStat: "max: <0.001s"}},
		{model.ConnectionItem{ConnCount: 1}, &model.Edge{MainStat: "avg: N/A", SecondaryStat: "max: N/A"}},
	}

	duration := &Duration{}

	for _, test := range tests {
		t.Run(test.want.Id, func(t *testing.T) {
			t.Parallel()

			edge := &model.Edge{}
			duration.FillEdgeStats(edge, test.ConnectionItem)

			assert.EqualValues(t, test.want, edge)
		},
		)
	}
}
