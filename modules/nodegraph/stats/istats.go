package stats

import (
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
)

type IStats interface {
	GetConfig() model.Config
	FillNodeStats(node *model.Node, connEndpoint model.ConnectionEndpoint)
	FillEdgeStats(edge *model.Edge, connItem model.ConnectionItem)
}
