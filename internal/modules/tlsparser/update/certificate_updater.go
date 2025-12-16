package update

import (
	ebpf_tools "github.com/k8spacket/k8spacket/internal/ebpf/tools"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grantae/certinfo"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/prometheus"
	"github.com/k8spacket/k8spacket/internal/thirdparty/network"
)

type CertificateUpdater struct {
	Network network.ConnectionInspector
}

func (updater *CertificateUpdater) Update(newValue *model.TLSDetails, oldValue *model.TLSDetails) {
	duration, _ := time.ParseDuration(os.Getenv("K8S_PACKET_TLS_CERTIFICATE_CACHE_TTL"))
	// do update when it is the first time or time to live is exceeded
	if !oldValue.Certificate.LastScrape.IsZero() && oldValue.Certificate.LastScrape.Add(duration).After(time.Now()) {
		newValue.Certificate = oldValue.Certificate
		return
	}
	scrapeCertificate(updater, newValue)
	ebpf_tools.StoreDomain(newValue.Dst, newValue.Port, newValue.Domain)

	if !newValue.Certificate.NotAfter.IsZero() {
		prometheus.K8sPacketTLSCertificateExpirationMetric.WithLabelValues(
			newValue.Dst,
			strconv.Itoa(int(newValue.Port)),
			newValue.Domain).Set(float64(newValue.Certificate.NotAfter.UnixMilli()))
	}
}

func scrapeCertificate(updater *CertificateUpdater, tlsDetails *model.TLSDetails) {
	domain := tlsDetails.Domain
	dst := tlsDetails.Dst
	port := tlsDetails.Port
	if port <= 0 {
		slog.Info("[updater scraping] dstPort is empty",
			"domain", domain,
			"dst", dst,
			"port", port,
			"Gave up", "")
		tlsDetails.Certificate.ServerChain = "UNAVAILABLE"
		tlsDetails.Certificate.LastScrape = time.Now()
		return
	}
	// check if domain is valid, if not - use destination IP
	if len(domain) <= 0 || !updater.Network.IsDomainReachable(domain) {
		domain = dst
	}

	certs, err := updater.Network.GetPeerCertificates(domain, port)
	if err != nil {
		slog.Error("[updater scraping] Error in Dial",
			"domain", domain,
			"port", port,
			"Trying with the default port...", "")
		port = 443
		certs, err = updater.Network.GetPeerCertificates(domain, port)
		if err != nil {
			slog.Error("[updater scraping] Error in Dial",
				"domain", domain,
				"port", port,
				"Gave up", "")
			tlsDetails.Certificate.ServerChain = "UNAVAILABLE"
			tlsDetails.Certificate.LastScrape = time.Now()
			return
		}
	}

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
	slog.Info("[updater scraping] TLS updater scraped",
		"domain", domain,
		"port", port)
}
