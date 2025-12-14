package broker

import (
	"github.com/k8spacket/k8spacket/pkg/events"
)

type IBroker interface {
	DistributeEvents()
	TCPEvent(event events.TCPEvent)
	TLSEvent(event events.TLSEvent)
}
