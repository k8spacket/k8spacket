package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strconv"
)

var (
	K8sPacketBytesSentMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_packet_bytes_sent",
			Help: "Kubernetes packet bytes sent",
		},
		[]string{"ns", "src", "src_name", "src_port", "dst", "dst_name", "dst_port", "persistent"},
	)
	K8sPacketBytesReceivedMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_packet_bytes_received",
			Help: "Kubernetes packet bytes received",
		},
		[]string{"ns", "src", "src_name", "src_port", "dst", "dst_name", "dst_port", "persistent"},
	)
	K8sPacketDurationSecondsMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_packet_duration_seconds",
			Help: "Kubernetes packet duration seconds",
		},
		[]string{"ns", "src", "src_name", "src_port", "dst", "dst_name", "dst_port", "persistent"},
	)
)

func init() {
	sendTCPMetrics, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_TCP_METRICS_ENABLED"))
	if sendTCPMetrics {
		prometheus.MustRegister(K8sPacketBytesSentMetric)
		prometheus.MustRegister(K8sPacketBytesReceivedMetric)
		prometheus.MustRegister(K8sPacketDurationSecondsMetric)
	}
}
