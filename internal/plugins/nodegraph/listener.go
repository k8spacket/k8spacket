package nodegraph

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/k8spacket/k8spacket/internal/plugins/nodegraph/prometheus"
	"github.com/k8spacket/k8spacket/pkg/events"
)

type Listener struct {
	service IService
}

func (listener *Listener) Listen(event events.TCPEvent) {

	var persistent = false
	var persistentDuration, _ = time.ParseDuration(os.Getenv("K8S_PACKET_TCP_PERSISTENT_DURATION"))
	if int(event.DeltaUs) > int(persistentDuration.Milliseconds()) {
		persistent = true
	}

	listener.service.recordConnection(ConnectionUpdate{
		Src:           event.Client.Addr,
		SrcName:       event.Client.Name,
		SrcNamespace:  event.Client.Namespace,
		Dst:           event.Server.Addr,
		DstName:       event.Server.Name,
		DstNamespace:  event.Server.Namespace,
		Persistent:    persistent,
		BytesSent:     float64(event.TxB),
		BytesReceived: float64(event.RxB),
		Duration:      float64(event.DeltaUs),
		Closed:        event.Closed,
	})

	if event.Closed {
		sendPrometheusMetrics(event, persistent)
		slog.Info("Connection",
			"src", event.Client.Addr,
			"srcName", event.Client.Name,
			"srcPort", strconv.Itoa(int(event.Client.Port)),
			"srcNS", event.Client.Namespace,
			"dst", event.Server.Addr,
			"dstName", event.Server.Name,
			"dstPort", strconv.Itoa(int(event.Server.Port)),
			"dstNS", event.Server.Namespace,
			"persistent", persistent,
			"bytesSent", float64(event.TxB),
			"bytesReceived", float64(event.RxB),
			"duration", float64(event.DeltaUs))
	}
}

func sendPrometheusMetrics(event events.TCPEvent, persistent bool) {
	hideSrcPort, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_TCP_METRICS_HIDE_SRC_PORT"))
	var srcPortMetrics = strconv.Itoa(int(event.Client.Port))
	if hideSrcPort {
		srcPortMetrics = "dynamic"
	}
	prometheus.K8sPacketBytesSentMetric.WithLabelValues(event.Client.Namespace, event.Client.Addr, event.Client.Name, srcPortMetrics, event.Server.Addr, event.Server.Name, strconv.Itoa(int(event.Server.Port)), strconv.FormatBool(persistent)).Observe(float64(event.TxB))
	prometheus.K8sPacketBytesReceivedMetric.WithLabelValues(event.Client.Namespace, event.Client.Addr, event.Client.Name, srcPortMetrics, event.Server.Addr, event.Server.Name, strconv.Itoa(int(event.Server.Port)), strconv.FormatBool(persistent)).Observe(float64(event.RxB))
	prometheus.K8sPacketDurationSecondsMetric.WithLabelValues(event.Client.Namespace, event.Client.Addr, event.Client.Name, srcPortMetrics, event.Server.Addr, event.Server.Name, strconv.Itoa(int(event.Server.Port)), strconv.FormatBool(persistent)).Observe(float64(event.DeltaUs))
}
