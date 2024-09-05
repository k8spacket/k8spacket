package broker

import (
	"github.com/k8spacket/k8spacket/modules"
)

type IBroker interface {
	DistributeEvents()
	TCPEvent(event modules.TCPEvent)
	TLSEvent(event modules.TLSEvent)
}
