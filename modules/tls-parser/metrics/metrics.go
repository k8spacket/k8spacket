package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/k8spacket/k8spacket/modules"
	"github.com/k8spacket/k8spacket/modules/idb"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/certificate"
	tls_connection_db "github.com/k8spacket/k8spacket/modules/tls-parser/metrics/db/tls_connection"
	tls_detail_db "github.com/k8spacket/k8spacket/modules/tls-parser/metrics/db/tls_detail"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/dict"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/prometheus"
	"strconv"
	"time"
)

func StoreTLSMetrics(tlsEvent modules.TLSEvent) {
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

	storeInDatabase(&tlsConnection, &tlsDetails)

	sendPrometheusMetrics(tlsConnection, tlsDetails)

	var j, _ = json.Marshal(tlsConnection)
	tls_parser_log.LOGGER.Println("TLS Record:", string(j))
}

func storeInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails) {
	var id = strconv.Itoa(int(idb.HashId(fmt.Sprintf("%s-%s", tlsConnection.Src, tlsConnection.Dst))))
	tlsConnection.Id = id
	tls_connection_db.Upsert(id, tlsConnection)
	tlsDetails.Id = id
	tls_detail_db.Upsert(id, tlsDetails, certificate.UpdateCertificateInfo)
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
