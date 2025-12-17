package ebpf_tc

import (
	"encoding/binary"
	"os"
	"testing"

	ebpf_tools "github.com/k8spacket/k8spacket/internal/ebpf/tools"
	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

type fakeBrokerTC struct {
	last modules.TLSEvent
}

func (f *fakeBrokerTC) DistributeEvents()               {}
func (f *fakeBrokerTC) TCPEvent(event modules.TCPEvent) {}
func (f *fakeBrokerTC) TLSEvent(event modules.TLSEvent) { f.last = event }

func TestDistribute(t *testing.T) {
	// ensure k8s enrichment uses disabled mode
	os.Setenv("K8S_PACKET_K8S_RESOURCES_DISABLED", "true")

	sNum := binary.BigEndian.Uint32([]byte{192, 168, 1, 100})
	dNum := binary.BigEndian.Uint32([]byte{10, 1, 2, 3})

	var evt tcTlsHandshakeEvent
	evt.Saddr = sNum
	evt.Daddr = dNum
	evt.Sport = 15000
	evt.Dport = 443
	evt.TlsVersion = 0x0304
	evt.TlsVersionsLength = 2 * 2 // two uint16
	evt.TlsVersions[0] = 0x0303
	evt.TlsVersions[1] = 0x0304
	evt.CiphersLength = 2 * 2
	evt.Ciphers[0] = 0x1301
	evt.Ciphers[1] = 0x1302
	name := "tc.example"
	copy(evt.ServerName[:], []byte(name))
	evt.ServerNameLength = uint16(len(name))
	evt.UsedTlsVersion = 0x0304
	evt.UsedCipher = 0x1301

	fb := &fakeBrokerTC{}
	tcInst := &EbpfTc{Broker: fb}

	distribute(evt, tcInst)

	got := fb.last
	assert.Equal(t, modules.TC, got.Source)
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
	assert.Equal(t, "N/A", got.Client.Name)
	assert.Equal(t, "N/A", got.Server.Name)
}
