package listener

import (
	"encoding/json"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/storer"
	"log/slog"
	"strconv"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/dict"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/prometheus"
)

type TlsListener struct {
	storer storer.Storer
}

func NewListener(storer storer.Storer) modules.Listener[modules.TLSEvent] {
	return &TlsListener{storer: storer}
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

	sendPrometheusMetrics(tlsConnection, tlsDetails)

	var j, _ = json.Marshal(tlsConnection)
	slog.Info("TLS connection", "Source", tlsEvent.Source.String(), "Record", string(j))
}

func sendPrometheusMetrics(tlsConnection model.TLSConnection, tlsDetails model.TLSDetails) {
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

	prometheus.K8sPacketTLSCertificateExpirationCounterMetric.WithLabelValues(
		tlsDetails.Dst,
		strconv.Itoa(int(tlsDetails.Port)),
		tlsDetails.Domain).Add(1)
}
