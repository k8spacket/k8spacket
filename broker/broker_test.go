package broker

import (
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/modules"
	"github.com/stretchr/testify/assert"
)

type mockNodegraphListener struct {
	modules.IListener[modules.TCPEvent]
	listenerCalled bool
}

func (mockNodegraphListener *mockNodegraphListener) Listen(event modules.TCPEvent) {
	mockNodegraphListener.listenerCalled = true
}

type mockTlsParserListener struct {
	modules.IListener[modules.TLSEvent]
	listenerCalled bool
}

func (mockTlsParserListener *mockTlsParserListener) Listen(event modules.TLSEvent) {
	mockTlsParserListener.listenerCalled = true
}

func TestDistributeEvents(t *testing.T) {

	mockNodegraphListener := &mockNodegraphListener{}
	mockTlsParserListener := &mockTlsParserListener{}

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
