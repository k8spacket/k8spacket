package broker

import (
	"github.com/k8spacket/k8spacket/plugins"
	"github.com/k8spacket/plugin-api"
)

var reassembledStreamChannel = make(chan plugin_api.ReassembledStream)
var tcpPacketPayoutChannel = make(chan plugin_api.TCPPacketPayload)

func ReassembledStreamEvent(stream plugin_api.ReassembledStream) {
	reassembledStreamChannel <- stream
}

func TCPPacketPayoutEvent(payload plugin_api.TCPPacketPayload) {
	tcpPacketPayoutChannel <- payload
}

func DistributeMessages(manager *plugins.Manager) {
	for {
		select {
		case message := <-reassembledStreamChannel:
			for _, plugin := range manager.GetStreamPlugins() {
				plugin.DistributeReassembledStream(message)
			}
		case message := <-tcpPacketPayoutChannel:
			for _, plugin := range manager.GetStreamPlugins() {
				plugin.DistributeTCPPacketPayload(message)
			}
		}
	}
}
