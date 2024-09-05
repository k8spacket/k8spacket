package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"
)

type Network struct {
	INetwork
}

func (network *Network) IsDomainReachable(domain string) bool {
	_, err := net.LookupIP(domain)
	if err != nil {
		return false
	}
	return true
}

func (network *Network) GetPeerCertificates(address string, port uint16) ([]*x509.Certificate, error) {

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 500 * time.Millisecond}, "tcp", fmt.Sprintf("%s:%d", address, port), conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.ConnectionState().PeerCertificates, nil
}
