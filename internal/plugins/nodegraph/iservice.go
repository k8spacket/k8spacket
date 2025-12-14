package nodegraph

import (
	"net/http"
	"regexp"
	"time"

	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/model"
)

type IService interface {
	recordConnection(update ConnectionUpdate)
	getConnections(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem

	getO11yStatsConfig(statsType string) (string, error)
	buildO11yResponse(r *http.Request) (model.NodeGraph, error)
}
