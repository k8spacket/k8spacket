package ebpf_inet

import (
	"encoding/binary"
	"testing"

	ebpf_tools "github.com/k8spacket/k8spacket/internal/ebpf/tools"
	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

type fakeBroker struct {
	last modules.TCPEvent
}

func (f *fakeBroker) DistributeEvents()               {}
func (f *fakeBroker) TCPEvent(event modules.TCPEvent) { f.last = event }
func (f *fakeBroker) TLSEvent(event modules.TLSEvent) {}

func TestDistribute(t *testing.T) {
	// use private IPs so EnrichAddress will set Name to "N/A"
	sBytes := []byte{192, 168, 0, 1}
	dBytes := []byte{10, 0, 0, 2}
	sNum := binary.LittleEndian.Uint32(sBytes)
	dNum := binary.LittleEndian.Uint32(dBytes)

	evt := bpfEvent{
		Saddr:   sNum,
		Daddr:   dNum,
		Sport:   12345,
		Dport:   80,
		TxB:     1000,
		RxB:     2000,
		DeltaUs: 5000,
		Closed:  true,
	}

	fb := &fakeBroker{}
	inet := &EbpfInet{Broker: fb}

	distribute(evt, inet)

	got := fb.last

	assert.Equal(t, ebpf_tools.IntToIP4(sNum, binary.LittleEndian.PutUint32), got.Client.Addr)
	assert.Equal(t, uint16(12345), got.Client.Port)
	assert.Equal(t, ebpf_tools.IntToIP4(dNum, binary.LittleEndian.PutUint32), got.Server.Addr)
	assert.Equal(t, uint16(80), got.Server.Port)
	assert.Equal(t, uint64(1000), got.TxB)
	assert.Equal(t, uint64(2000), got.RxB)
	// DeltaUs is divided by 1000 in distribute
	assert.Equal(t, uint64(5), got.DeltaUs)
	assert.True(t, got.Closed)
	// EnrichAddress for private IPs should set Name to "N/A"
	assert.Equal(t, "N/A", got.Client.Name)
	assert.Equal(t, "N/A", got.Server.Name)
}
