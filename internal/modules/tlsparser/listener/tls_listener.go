package listener

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/storer"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/dict"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/prometheus"
)

type TlsListener struct {
	storer                      storer.Storer
	tlsRecordsMeticsEnabled     bool
	tlsExpirationMetricsEnabled bool
}

func NewListener(storer storer.Storer) modules.Listener[modules.TLSEvent] {
	tlsRecordsMeticsEnabled, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_TLS_RECORDS_METRICS_ENABLED"))
	tlsExpirationMetricsEnabled, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_TLS_EXPIRATION_METRICS_ENABLED"))
	return &TlsListener{storer: storer,
		tlsRecordsMeticsEnabled:     tlsRecordsMeticsEnabled,
		tlsExpirationMetricsEnabled: tlsExpirationMetricsEnabled,
	}
}

func (listener *TlsListener) Listen(tlsEvent modules.TLSEvent) {

	tlsConnection := model.TLSConnection{
		Src:             tlsEvent.Client.Addr,
		SrcName:         tlsEvent.Client.Name,
		SrcNamespace:    tlsEvent.Client.Namespace,
		Dst:             tlsEvent.Server.Addr,
		DstName:         tlsEvent.Server.Name,
		DstPort:         tlsEvent.Server.Port,
		Domain:          tlsEvent.ServerName,
		UsedTLSVersion:  dict.ParseTLSVersion(tlsEvent.UsedTlsVersion),
		UsedCipherSuite: dict.ParseCipherSuite(tlsEvent.UsedCipher),
		LastSeen:        time.Now()}

	tlsDetails := model.TLSDetails{
		Domain:          tlsEvent.ServerName,
		Dst:             tlsEvent.Server.Addr,
		Port:            tlsEvent.Server.Port,
		UsedTLSVersion:  dict.ParseTLSVersion(tlsEvent.UsedTlsVersion),
		UsedCipherSuite: dict.ParseCipherSuite(tlsEvent.UsedCipher)}

	for _, tlsVersion := range tlsEvent.TlsVersions {
		tlsDetails.ClientTLSVersions = append(tlsDetails.ClientTLSVersions, dict.ParseTLSVersion(tlsVersion))
	}
	for _, cipher := range tlsEvent.Ciphers {
		tlsDetails.ClientCipherSuites = append(tlsDetails.ClientCipherSuites, dict.ParseCipherSuite(cipher))
	}

	listener.storer.StoreInDatabase(&tlsConnection, &tlsDetails)

	sendPrometheusMetrics(tlsConnection, tlsDetails, listener.tlsRecordsMeticsEnabled, listener.tlsExpirationMetricsEnabled)

	var j, _ = json.Marshal(tlsConnection)
	slog.Info("TLS connection", "Source", tlsEvent.Source.String(), "Record", string(j))
}

func sendPrometheusMetrics(tlsConnection model.TLSConnection, tlsDetails model.TLSDetails, tlsRecordsMeticsEnabled bool, tlsExpirationMetricsEnabled bool) {
	if tlsRecordsMeticsEnabled {
		prometheus.K8sPacketTLSRecordMetric.WithLabelValues(
			tlsConnection.SrcNamespace,
			tlsConnection.Src,
			tlsConnection.SrcName,
			tlsConnection.Dst,
			tlsConnection.DstName,
			strconv.Itoa(int(tlsConnection.DstPort)),
			tlsConnection.Domain,
			tlsConnection.UsedTLSVersion,
			tlsConnection.UsedCipherSuite).Add(1)
	}
	if tlsExpirationMetricsEnabled {
		prometheus.K8sPacketTLSCertificateExpirationCounterMetric.WithLabelValues(
			tlsDetails.Dst,
			strconv.Itoa(int(tlsDetails.Port)),
			tlsDetails.Domain).Add(1)
	}
}
