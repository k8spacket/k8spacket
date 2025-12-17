package ebpf_socketfilter

import (
	"encoding/binary"
	"testing"

	ebpf_tools "github.com/k8spacket/k8spacket/internal/ebpf/tools"
	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

type fakeBrokerSF struct {
	last modules.TLSEvent
}

func (f *fakeBrokerSF) DistributeEvents()               {}
func (f *fakeBrokerSF) TCPEvent(event modules.TCPEvent) {}
func (f *fakeBrokerSF) TLSEvent(event modules.TLSEvent) { f.last = event }


func TestDistribute(t *testing.T) {
	sNum := binary.BigEndian.Uint32([]byte{192, 168, 0, 10})
	dNum := binary.BigEndian.Uint32([]byte{10, 0, 0, 5})

	var evt socketfilterTlsHandshakeEvent
	evt.Saddr = sNum
	evt.Daddr = dNum
	evt.Sport = 44321
	evt.Dport = 443
	evt.TlsVersion = 0x0304
	// set TlsVersionsLength = 4 bytes -> two uint16 entries
	evt.TlsVersionsLength = 4
	evt.TlsVersions[0] = 0x0303
	evt.TlsVersions[1] = 0x0304
	// set ciphers length 4 -> two entries
	evt.CiphersLength = 4
	evt.Ciphers[0] = 0x1301
	evt.Ciphers[1] = 0x1302
	// server name
	name := "example.local"
	copy(evt.ServerName[:], []byte(name))
	evt.ServerNameLength = uint16(len(name))
	evt.UsedTlsVersion = 0x0304
	evt.UsedCipher = 0x1301

	fb := &fakeBrokerSF{}
	filter := &EbpfSocketFilter{Broker: fb}

	distribute(evt, filter)

	got := fb.last
	assert.Equal(t, modules.SocketFilter, got.Source)
	assert.Equal(t, ebpf_tools.IntToIP4(sNum, binary.BigEndian.PutUint32), got.Client.Addr)
	assert.Equal(t, evt.Sport, got.Client.Port)
	assert.Equal(t, ebpf_tools.IntToIP4(dNum, binary.BigEndian.PutUint32), got.Server.Addr)
	assert.Equal(t, evt.Dport, got.Server.Port)
	assert.Equal(t, 2, len(got.TlsVersions))
	assert.Equal(t, uint16(0x0303), got.TlsVersions[0])
	assert.Equal(t, uint16(0x0304), got.TlsVersions[1])
	assert.Equal(t, 2, len(got.Ciphers))
	assert.Equal(t, uint16(0x1301), got.Ciphers[0])
	assert.Equal(t, uint16(0x1302), got.Ciphers[1])
	assert.Equal(t, name, got.ServerName)
	assert.Equal(t, evt.UsedTlsVersion, got.UsedTlsVersion)
	assert.Equal(t, evt.UsedCipher, got.UsedCipher)
	// EnrichAddress for private IPs sets Name to "N/A"
	assert.Equal(t, "N/A", got.Client.Name)
	assert.Equal(t, "N/A", got.Server.Name)
}
