package broker

import (
	"github.com/k8spacket/k8spacket/modules"
)

type Broker struct {
	NodegraphListener modules.IListener[modules.TCPEvent]
	TlsParserListener modules.IListener[modules.TLSEvent]
}

var tcpEventChannel = make(chan modules.TCPEvent)
var tlsEventChannel = make(chan modules.TLSEvent)

func (broker *Broker) TCPEvent(event modules.TCPEvent) {
	tcpEventChannel <- event
}

func (broker *Broker) TLSEvent(event modules.TLSEvent) {
	tlsEventChannel <- event
}

func (broker *Broker) DistributeEvents() {
	for {
		select {
		case event := <-tcpEventChannel:
			broker.NodegraphListener.Listen(event)
		case event := <-tlsEventChannel:
			broker.TlsParserListener.Listen(event)
		}
	}
}
