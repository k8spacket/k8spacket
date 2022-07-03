package nodegraph

import (
	"github.com/k8spacket/metrics/nodegraph/model"
	"sync"
)

var (
	connectionItems      = make(map[string]model.ConnectionItem)
	connectionItemsMutex = sync.RWMutex{}
)

func UpdateNodeGraph(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, closed bool, bytesSent float64, bytesReceived float64, duration float64) {
	connectionItemsMutex.Lock()
	var connection = connectionItems[src+"-"+dst]
	if (model.ConnectionItem{} == connection) {
		connection = *&model.ConnectionItem{src, srcName, srcNamespace, dst, dstName, dstNamespace, 0, 0, 0, 0, 0, 0}
	}
	connection.ConnCount++
	if closed {
		connection.ConnClosed++
	}
	connection.BytesSent += bytesSent
	connection.BytesReceived += bytesReceived
	connection.Duration += duration
	if duration > connection.MaxDuration {
		connection.MaxDuration = duration
	}
	connectionItems[src+"-"+dst] = connection
	connectionItemsMutex.Unlock()
}
