package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	K8sPacketBytesSentMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_packet_bytes_sent",
			Help: "Kubernetes packet bytes sent",
		},
		[]string{"ns", "src", "src_name", "src_port", "dst", "dst_name", "dst_port", "closed"},
	)
	K8sPacketBytesReceivedMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_packet_bytes_received",
			Help: "Kubernetes packet bytes received",
		},
		[]string{"ns", "src", "src_name", "src_port", "dst", "dst_name", "dst_port", "closed"},
	)
	K8sPacketDurationSecondsMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "k8s_packet_duration_seconds",
			Help: "Kubernetes packet duration seconds",
		},
		[]string{"ns", "src", "src_name", "src_port", "dst", "dst_name", "dst_port", "closed"},
	)
)

func init() {
	prometheus.MustRegister(K8sPacketBytesSentMetric)
	prometheus.MustRegister(K8sPacketBytesReceivedMetric)
	prometheus.MustRegister(K8sPacketDurationSecondsMetric)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
}
