package nodegraph

import (
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"net/http"
	"regexp"
	"time"
)

type IService interface {
	update(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64)
	getConnections(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem

	getO11yStatsConfig(statsType string) (string, error)
	buildO11yResponse(r *http.Request) (model.NodeGraph, error)
}
