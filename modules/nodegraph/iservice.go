package nodegraph

import (
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"net/http"
	"regexp"
	"time"
)

type IService interface {
	Update(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64)
	GetConnections(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem

	GetO11yStatsConfig(r *http.Request) (string, error)
	BuildO11yResponse(r *http.Request) (model.NodeGraph, error)
}
