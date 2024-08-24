package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strconv"
)

var (
	K8sPacketTLSRecordMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_packet_tls_record",
			Help: "Kubernetes packet TLS Record",
		},
		[]string{"ns", "src", "src_name", "dst", "dst_name", "dst_port", "domain", "tls_version", "cipher_suite"},
	)

	K8sPacketTLSCertificateExpirationCounterMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_packet_tls_cert_expiry_count",
			Help: "Kubernetes packet TLS certificate expiration counter",
		},
		[]string{"dst", "dst_port", "domain"},
	)
	K8sPacketTLSCertificateExpirationMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_packet_tls_cert_expiry",
			Help: "Kubernetes packet TLS certificate expiration",
		},
		[]string{"dst", "dst_port", "domain"},
	)
)

func init() {
	sendTLSMetrics, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_TLS_METRICS_ENABLED"))
	if sendTLSMetrics {
		prometheus.MustRegister(K8sPacketTLSRecordMetric)
		prometheus.MustRegister(K8sPacketTLSCertificateExpirationMetric)
		prometheus.MustRegister(K8sPacketTLSCertificateExpirationCounterMetric)
	}
}
