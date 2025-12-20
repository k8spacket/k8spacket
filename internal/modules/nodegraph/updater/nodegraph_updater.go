package updater

import (
	"fmt"
	"strconv"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/repository"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
)

type NodegraphUpdater struct {
	repo repository.Repository[model.ConnectionItem]
}

func NewUpdater(repo repository.Repository[model.ConnectionItem]) *NodegraphUpdater {
	return &NodegraphUpdater{repo: repo}
}

func (updater *NodegraphUpdater) Update(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64, closed bool) {
	var id = strconv.Itoa(int(db.HashId(fmt.Sprintf("%s-%s", src, dst))))
	var connection = updater.repo.Read(id)
	if (model.ConnectionItem{} == connection) {
		connection = model.ConnectionItem{Src: src, Dst: dst}
	}
	connection.SrcName = srcName
	connection.SrcNamespace = srcNamespace
	connection.DstName = dstName
	connection.DstNamespace = dstNamespace
	if closed {
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
	}
	connection.LastSeen = time.Now()
	updater.repo.Set(id, &connection)
}
