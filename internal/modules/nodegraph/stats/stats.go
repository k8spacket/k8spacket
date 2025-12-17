package stats

import (
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
)

type Stats interface {
	GetConfig() model.Config
	FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint)
	FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem)
}
