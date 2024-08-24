package broker

import (
	"github.com/k8spacket/k8spacket/modules"
	tcp_metrics "github.com/k8spacket/k8spacket/modules/nodegraph/metrics"
	tls_metrics "github.com/k8spacket/k8spacket/modules/tls-parser/metrics"
)

var tcpEventChannel = make(chan modules.TCPEvent)
var tlsEventChannel = make(chan modules.TLSEvent)

func TCPEvent(event modules.TCPEvent) {
	tcpEventChannel <- event
}

func TLSEvent(event modules.TLSEvent) {
	tlsEventChannel <- event
}

func DistributeEvents() {
	for {
		select {
		case event := <-tcpEventChannel:
			tcp_metrics.StoreNodegraphMetric(event)
		case event := <-tlsEventChannel:
			tls_metrics.StoreTLSMetrics(event)
		}
	}
}
