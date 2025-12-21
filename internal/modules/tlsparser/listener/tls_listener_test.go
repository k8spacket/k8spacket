package listener

import (
	"bytes"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/storer"
	"log/slog"
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

type mockStorer struct {
	storer.Storer
	client, server, domain, usedCipher string
	clientTLSVersions                  []string
}

func (mock *mockStorer) StoreInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails) {
	mock.client = tlsConnection.Src
	mock.server = tlsConnection.Dst
	mock.domain = tlsConnection.Domain
	mock.usedCipher = tlsConnection.UsedCipherSuite
	mock.clientTLSVersions = tlsDetails.ClientTLSVersions
}

func TestListen(t *testing.T) {

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	mockStorer := &mockStorer{}
	listener := NewListener(mockStorer)

	event := modules.TLSEvent{Client: modules.Address{Addr: "client"},
		Server:      modules.Address{Addr: "server"},
		ServerName:  "k8spacket.io",
		TlsVersions: []uint16{0x0303, 0x0302}, UsedTlsVersion: 0x0303,
		Ciphers: []uint16{0x0024, 0x0009, 0x000C}, UsedCipher: 0x0024}
	listener.Listen(event)

	assert.EqualValues(t, event.Client.Addr, mockStorer.client)
	assert.EqualValues(t, event.Server.Addr, mockStorer.server)
	assert.EqualValues(t, event.ServerName, mockStorer.domain)
	assert.EqualValues(t, "TLS_KRB5_WITH_RC4_128_MD5", mockStorer.usedCipher)
	assert.EqualValues(t, []string{"TLS 1.2", "TLS 1.1"}, mockStorer.clientTLSVersions)

	assert.Contains(t, str.String(), "TLS connection")

}
