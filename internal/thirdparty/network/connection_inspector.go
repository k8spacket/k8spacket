package network

import "crypto/x509"

type ConnectionInspector interface {
	IsDomainReachable(domain string) bool
	GetPeerCertificates(address string, port uint16) ([]*x509.Certificate, error)
}
