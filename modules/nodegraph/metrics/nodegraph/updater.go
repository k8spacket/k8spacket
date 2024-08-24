package nodegraph

import (
	"fmt"
	"github.com/k8spacket/k8spacket/modules/idb"
	tcp_connection_db "github.com/k8spacket/k8spacket/modules/nodegraph/metrics/db/tcp_connection"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph/model"
	"strconv"
	"sync"
	"time"
)

var connectionItemsMutex = sync.RWMutex{}

func UpdateNodeGraph(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64) {
	connectionItemsMutex.Lock()
	var id = strconv.Itoa(int(idb.HashId(fmt.Sprintf("%s-%s", src, dst))))
	var connection = tcp_connection_db.Read(id)
	if (model.ConnectionItem{} == connection) {
		connection = *&model.ConnectionItem{Src: src, Dst: dst}
	}
	connection.SrcName = srcName
	connection.SrcNamespace = srcNamespace
	connection.DstName = dstName
	connection.DstNamespace = dstNamespace
	connection.ConnCount++
	if persistent {
		connection.ConnPersistent++
	}
	connection.BytesSent += bytesSent
	connection.BytesReceived += bytesReceived
	connection.Duration += duration
	if duration > connection.MaxDuration {
		connection.MaxDuration = duration
	}
	connection.LastSeen = time.Now()
	tcp_connection_db.Set(id, &connection)
	connectionItemsMutex.Unlock()
}
