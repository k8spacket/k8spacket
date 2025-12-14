package nodegraph

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

func (mockService *mockService) update(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64, closed bool) {
	mockService.client = src
	mockService.server = dst
}

func TestListen(t *testing.T) {

	var str bytes.Buffer

	os.Setenv("K8S_PACKET_TCP_PERSISTENT_DURATION", "1ms")
	os.Setenv("K8S_PACKET_TCP_METRICS_HIDE_SRC_PORT", "true")

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	service := &mockService{}
	listener := &Listener{service}

	event := modules.TCPEvent{Client: modules.Address{Addr: "client"}, Server: modules.Address{Addr: "server"}, DeltaUs: 2, Closed: true}
	listener.Listen(event)

	assert.EqualValues(t, event.Client.Addr, service.client)
	assert.EqualValues(t, event.Server.Addr, service.server)

	assert.Contains(t, str.String(), "Connection src=client srcName=\"\" srcPort=0 srcNS=\"\" dst=server dstName=\"\" dstPort=0 dstNS=\"\" persistent=true bytesSent=0 bytesReceived=0 duration=2")

}
