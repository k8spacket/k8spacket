package broker

import (
	"github.com/k8spacket/k8spacket/pkg/api"
	"github.com/k8spacket/k8spacket/pkg/events"
)

type Broker struct {
	IBroker
	NodegraphListener api.IListener[events.TCPEvent]
	TlsParserListener api.IListener[events.TLSEvent]
	tcpEventChannel   chan events.TCPEvent
	tlsEventChannel   chan events.TLSEvent
}

func Init(nodegraphListener api.IListener[events.TCPEvent], tlsParserListener api.IListener[events.TLSEvent]) *Broker {
	broker := Broker{NodegraphListener: nodegraphListener, TlsParserListener: tlsParserListener}
	broker.tcpEventChannel = make(chan events.TCPEvent)
	broker.tlsEventChannel = make(chan events.TLSEvent)
	return &broker
}

func (broker *Broker) TCPEvent(event events.TCPEvent) {
	broker.tcpEventChannel <- event
}

func (broker *Broker) TLSEvent(event events.TLSEvent) {
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
