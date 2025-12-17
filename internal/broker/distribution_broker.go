package broker

import (
	"github.com/k8spacket/k8spacket/internal/modules"
)

type DistributionBroker struct {
	Broker
	NodegraphListener modules.Listener[modules.TCPEvent]
	TlsParserListener modules.Listener[modules.TLSEvent]
	tcpEventChannel   chan modules.TCPEvent
	tlsEventChannel   chan modules.TLSEvent
}

func Init(nodegraphListener modules.Listener[modules.TCPEvent], tlsParserListener modules.Listener[modules.TLSEvent]) *DistributionBroker {
	broker := DistributionBroker{NodegraphListener: nodegraphListener, TlsParserListener: tlsParserListener}
	broker.tcpEventChannel = make(chan modules.TCPEvent)
	broker.tlsEventChannel = make(chan modules.TLSEvent)
	return &broker
}

func (broker *DistributionBroker) TCPEvent(event modules.TCPEvent) {
	broker.tcpEventChannel <- event
}

func (broker *DistributionBroker) TLSEvent(event modules.TLSEvent) {
	broker.tlsEventChannel <- event
}

func (broker *DistributionBroker) DistributeEvents() {
	for {
		select {
		case event := <-broker.tcpEventChannel:
			broker.NodegraphListener.Listen(event)
		case event := <-broker.tlsEventChannel:
			broker.TlsParserListener.Listen(event)
		}
	}
}
