package broker

import (
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

type mockTcpListener struct {
	modules.Listener[modules.TCPEvent]
	listenerCalled bool
}

func (mockTcpListener *mockTcpListener) Listen(event modules.TCPEvent) {
	mockTcpListener.listenerCalled = true
}

type mockTlsListener struct {
	modules.Listener[modules.TLSEvent]
	listenerCalled bool
}

func (mockTlsListener *mockTlsListener) Listen(event modules.TLSEvent) {
	mockTlsListener.listenerCalled = true
}

func TestDistributeEvents(t *testing.T) {

	mockNodegraphListener := &mockTcpListener{}
	mockTlsParserListener := &mockTlsListener{}

	broker := Init(mockNodegraphListener, mockTlsParserListener)

	go broker.DistributeEvents()

	assert.EqualValues(t, false, mockNodegraphListener.listenerCalled)

	assert.EqualValues(t, false, mockTlsParserListener.listenerCalled)

	broker.TCPEvent(modules.TCPEvent{Client: modules.Address{Addr: "addr1"}, TxB: 100})

	broker.TLSEvent(modules.TLSEvent{Client: modules.Address{Addr: "addr1"}, ServerName: "k8spacket.io"})

	assert.Eventually(t, func() bool {
		return mockNodegraphListener.listenerCalled && mockTlsParserListener.listenerCalled
	}, time.Second*1, time.Millisecond*100)

}
