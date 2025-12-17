package broker

import (
	"github.com/k8spacket/k8spacket/internal/modules"
)

type Broker interface {
	DistributeEvents()
	TCPEvent(event modules.TCPEvent)
	TLSEvent(event modules.TLSEvent)
}
