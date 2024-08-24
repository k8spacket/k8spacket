package certificate

import (
	"crypto/tls"
	"fmt"
	"github.com/grantae/certinfo"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/prometheus"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func UpdateCertificateInfo(newValue *model.TLSDetails, oldValue *model.TLSDetails) {
	duration, _ := time.ParseDuration(os.Getenv("K8S_PACKET_TLS_CERTIFICATE_CACHE_TTL"))
	// do update when it is the first time or time to live is exceeded
	if !oldValue.Certificate.LastScrape.IsZero() && oldValue.Certificate.LastScrape.Add(duration).After(time.Now()) {
		newValue.Certificate = oldValue.Certificate
		return
	}
	scrapeCertificate(newValue)

	if !newValue.Certificate.NotAfter.IsZero() {
		prometheus.K8sPacketTLSCertificateExpirationMetric.WithLabelValues(
			newValue.Dst,
			strconv.Itoa(int(newValue.Port)),
			newValue.Domain).Set(float64(newValue.Certificate.NotAfter.UnixMilli()))
	}
}

func scrapeCertificate(tlsDetails *model.TLSDetails) {
	domain := tlsDetails.Domain
	dst := tlsDetails.Dst
	port := tlsDetails.Port
	if port <= 0 {
		tls_parser_log.LOGGER.Printf("[certificate scraping] dstPort is empty, domain: %s, dst: %s, port %d. Gave up", domain, dst, port)
		tlsDetails.Certificate.ServerChain = "UNAVAILABLE"
		tlsDetails.Certificate.LastScrape = time.Now()
		return
	}
	// check if domain is valid, if not use destination IP
	if len(domain) > 0 {
		_, err := net.LookupIP(domain)
		if err != nil {
			domain = dst
		}
	} else {
		domain = dst
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 500 * time.Millisecond}, "tcp", fmt.Sprintf("%s:%d", domain, port), conf)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[certificate scraping] Error in Dial, domain: %s, port %d. Trying with the default port...", domain, port)
		port = 443
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 500 * time.Millisecond}, "tcp", fmt.Sprintf("%s:%d", domain, port), conf)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[certificate scraping] Error in Dial, domain: %s, port %d. Gave up", domain, port)
			tlsDetails.Certificate.ServerChain = "UNAVAILABLE"
			tlsDetails.Certificate.LastScrape = time.Now()
			return
		}
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	chain := ""
	for _, cert := range certs {
		if cert == certs[0] {
			tlsDetails.Certificate.NotBefore = cert.NotBefore
			tlsDetails.Certificate.NotAfter = cert.NotAfter
			tlsDetails.Certificate.LastScrape = time.Now()

		}
		certString, _ := certinfo.CertificateText(cert)
		chain += strings.Replace(certString, "\n\n", "\n", -1)
	}
	tlsDetails.Certificate.ServerChain = chain
	tls_parser_log.LOGGER.Printf("[certificate scraping] TLS certificate scraped, domain: %s, port %d", domain, port)
}
