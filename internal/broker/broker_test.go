package broker

import (
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/pkg/api"
	"github.com/k8spacket/k8spacket/pkg/events"
	"github.com/stretchr/testify/assert"
)

type mockNodegraphListener struct {
	api.IListener[events.TCPEvent]
	listenerCalled bool
}

func (mockNodegraphListener *mockNodegraphListener) Listen(event events.TCPEvent) {
	mockNodegraphListener.listenerCalled = true
}

type mockTlsParserListener struct {
	api.IListener[events.TLSEvent]
	listenerCalled bool
}

func (mockTlsParserListener *mockTlsParserListener) Listen(event events.TLSEvent) {
	mockTlsParserListener.listenerCalled = true
}

func TestDistributeEvents(t *testing.T) {

	mockNodegraphListener := &mockNodegraphListener{}
	mockTlsParserListener := &mockTlsParserListener{}

	broker := Init(mockNodegraphListener, mockTlsParserListener)

	go broker.DistributeEvents()

	assert.EqualValues(t, false, mockNodegraphListener.listenerCalled)

	assert.EqualValues(t, false, mockTlsParserListener.listenerCalled)

	broker.TCPEvent(events.TCPEvent{Client: events.Address{Addr: "addr1"}, TxB: 100})

	broker.TLSEvent(events.TLSEvent{Client: events.Address{Addr: "addr1"}, ServerName: "k8spacket.io"})

	assert.Eventually(t, func() bool {
		return mockNodegraphListener.listenerCalled && mockTlsParserListener.listenerCalled
	}, time.Second*1, time.Millisecond*100)

}
