package broker

import (
	"github.com/k8spacket/k8spacket/plugins"
	"github.com/k8spacket/plugin-api/v2"
)

var tcpEventChannel = make(chan plugin_api.TCPEvent)
var tlsEventChannel = make(chan plugin_api.TLSEvent)

func TCPEvent(event plugin_api.TCPEvent) {
	tcpEventChannel <- event
}

func TLSEvent(event plugin_api.TLSEvent) {
	tlsEventChannel <- event
}

func DistributeEvents(manager *plugins.Manager) {
	for {
		select {
		case event := <-tcpEventChannel:
			for _, plugin := range manager.GetTCPConsumerPlugins() {
				plugin.DistributeTCPEvent(event)
			}
		case event := <-tlsEventChannel:
			for _, plugin := range manager.GetTLSConsumerPlugins() {
				plugin.DistributeTLSEvent(event)
			}
		}
	}
}
