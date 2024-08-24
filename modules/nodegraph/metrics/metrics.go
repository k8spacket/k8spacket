package metrics

import (
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/nodegraph/log"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/prometheus"
	"os"
	"strconv"
	"time"
)

func StoreNodegraphMetric(event modules.TCPEvent) {

	var persistent = false
	var persistentDuration, _ = time.ParseDuration(os.Getenv("K8S_PACKET_TCP_PERSISTENT_DURATION"))
	if int(event.DeltaUs) > int(persistentDuration.Milliseconds()) {
		persistent = true
	}

	sendPrometheusMetrics(event, persistent)

	nodegraph.UpdateNodeGraph(event.Client.Addr, event.Client.Name, event.Client.Namespace, event.Server.Addr, event.Server.Name, event.Server.Namespace, persistent, float64(event.TxB), float64(event.RxB), float64(event.DeltaUs))

	nodegraph_log.LOGGER.Printf("Connection: src=%v srcName=%v srcPort=%v srcNS=%v dst=%v dstName=%v dstPort=%v dstNS=%v persistent=%v bytesSent=%v bytesReceived=%v duration=%v",
		event.Client.Addr,
		event.Client.Name,
		strconv.Itoa(int(event.Client.Port)),
		event.Client.Namespace,
		event.Server.Addr,
		event.Server.Name,
		strconv.Itoa(int(event.Server.Port)),
		event.Server.Namespace,
		persistent,
		float64(event.TxB),
		float64(event.RxB),
		float64(event.DeltaUs))
}

func sendPrometheusMetrics(event modules.TCPEvent, persistent bool) {
	hideSrcPort, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_TCP_METRICS_HIDE_SRC_PORT"))
	var srcPortMetrics = strconv.Itoa(int(event.Client.Port))
	if hideSrcPort {
		srcPortMetrics = "dynamic"
	}
	prometheus.K8sPacketBytesSentMetric.WithLabelValues(event.Client.Namespace, event.Client.Addr, event.Client.Name, srcPortMetrics, event.Server.Addr, event.Server.Name, strconv.Itoa(int(event.Server.Port)), strconv.FormatBool(persistent)).Observe(float64(event.TxB))
	prometheus.K8sPacketBytesReceivedMetric.WithLabelValues(event.Client.Namespace, event.Client.Addr, event.Client.Name, srcPortMetrics, event.Server.Addr, event.Server.Name, strconv.Itoa(int(event.Server.Port)), strconv.FormatBool(persistent)).Observe(float64(event.RxB))
	prometheus.K8sPacketDurationSecondsMetric.WithLabelValues(event.Client.Namespace, event.Client.Addr, event.Client.Name, srcPortMetrics, event.Server.Addr, event.Server.Name, strconv.Itoa(int(event.Server.Port)), strconv.FormatBool(persistent)).Observe(float64(event.DeltaUs))
}
