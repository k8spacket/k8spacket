package network

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDomainReachable(t *testing.T) {
	inspector := &HttpConnectionInspector{}

	ok := inspector.IsDomainReachable("localhost")
	assert.True(t, ok, "localhost should be reachable")

	notOk := inspector.IsDomainReachable("nonexistent.invalid")
	assert.False(t, notOk, "nonexistent.invalid should not resolve")
}

func loadCertFromFiles(t *testing.T) (tls.Certificate, []*x509.Certificate) {
	base := "../../../tests/units"
	keyPath := base + "/key.pem"
	chainPath := base + "/chain.pem"

	keyPEM, err := os.ReadFile(keyPath)
	assert.NoError(t, err)
	chainPEM, err := os.ReadFile(chainPath)
	assert.NoError(t, err)

	tlsCert, err := tls.X509KeyPair(chainPEM, keyPEM)
	assert.NoError(t, err)

	// Parse all certificates from the DER blobs included in tlsCert.Certificate
	var parsed []*x509.Certificate
	for _, der := range tlsCert.Certificate {
		cs, err := x509.ParseCertificates(der)
		assert.NoError(t, err)
		parsed = append(parsed, cs...)
	}
	assert.NotEmpty(t, parsed)

	return tlsCert, parsed
}

func TestGetPeerCertificates(t *testing.T) {
	tlsCert, certsFromFiles := loadCertFromFiles(t)

	// start TLS listener on ephemeral port
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{tlsCert}})
	assert.NoError(t, err)
	defer ln.Close()

	// accept one connection (handshake) in background
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			if tc, ok := conn.(*tls.Conn); ok {
				_ = tc.Handshake()
			}
			conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)

	inspector := &HttpConnectionInspector{}

	certs, err := inspector.GetPeerCertificates("127.0.0.1", uint16(addr.Port))
	assert.NoError(t, err)
	assert.NotEmpty(t, certs)

	// Expect at least server cert and CA cert presented in the chain
	assert.GreaterOrEqual(t, len(certs), 2)
	assert.Equal(t, certsFromFiles[0].Subject.CommonName, certs[0].Subject.CommonName)
	assert.Equal(t, certsFromFiles[1].Subject.CommonName, certs[1].Subject.CommonName)
}
