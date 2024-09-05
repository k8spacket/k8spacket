package broker

import (
	"github.com/k8spacket/k8spacket/modules"
)

type Broker struct {
	IBroker
	NodegraphListener modules.IListener[modules.TCPEvent]
	TlsParserListener modules.IListener[modules.TLSEvent]
	tcpEventChannel   chan modules.TCPEvent
	tlsEventChannel   chan modules.TLSEvent
}

func Init(nodegraphListener modules.IListener[modules.TCPEvent], tlsParserListener modules.IListener[modules.TLSEvent]) *Broker {
	broker := Broker{NodegraphListener: nodegraphListener, TlsParserListener: tlsParserListener}
	broker.tcpEventChannel = make(chan modules.TCPEvent)
	broker.tlsEventChannel = make(chan modules.TLSEvent)
	return &broker
}

func (broker *Broker) TCPEvent(event modules.TCPEvent) {
	broker.tcpEventChannel <- event
}

func (broker *Broker) TLSEvent(event modules.TLSEvent) {
	broker.tlsEventChannel <- event
}

func (broker *Broker) DistributeEvents() {
	for {
		select {
		case event := <-broker.tcpEventChannel:
			broker.NodegraphListener.Listen(event)
		case event := <-broker.tlsEventChannel:
			broker.TlsParserListener.Listen(event)
		}
	}
}
