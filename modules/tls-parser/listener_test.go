package tlsparser

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/k8spacket/k8spacket/modules"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	service := &mockService{}
	listener := &Listener{service}

	event := modules.TLSEvent{Client: modules.Address{Addr: "client"},
		Server:      modules.Address{Addr: "server"},
		ServerName:  "k8spacket.io",
		TlsVersions: []uint16{0x0303, 0x0302}, UsedTlsVersion: 0x0303,
		Ciphers: []uint16{0x0024, 0x0009, 0x000C}, UsedCipher: 0x0024}
	listener.Listen(event)

	assert.EqualValues(t, event.Client.Addr, service.client)
	assert.EqualValues(t, event.Server.Addr, service.server)
	assert.EqualValues(t, event.ServerName, service.domain)
	assert.EqualValues(t, "TLS_KRB5_WITH_RC4_128_MD5", service.usedCipher)
	assert.EqualValues(t, []string{"TLS 1.2", "TLS 1.1"}, service.clientTLSVersions)

	assert.Contains(t, str.String(), "TLS connection")

}
