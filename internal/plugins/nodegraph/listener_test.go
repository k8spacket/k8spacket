package nodegraph

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/k8spacket/k8spacket/pkg/events"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {

	var str bytes.Buffer

	os.Setenv("K8S_PACKET_TCP_PERSISTENT_DURATION", "1ms")
	os.Setenv("K8S_PACKET_TCP_METRICS_HIDE_SRC_PORT", "true")

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	service := &mockService{}
	listener := &Listener{service}

	event := events.TCPEvent{Client: events.Address{Addr: "client"}, Server: events.Address{Addr: "server"}, DeltaUs: 2, Closed: true}
	listener.Listen(event)

	assert.EqualValues(t, event.Client.Addr, service.client)
	assert.EqualValues(t, event.Server.Addr, service.server)

	assert.Contains(t, str.String(), "Connection src=client srcName=\"\" srcPort=0 srcNS=\"\" dst=server dstName=\"\" dstPort=0 dstNS=\"\" persistent=true bytesSent=0 bytesReceived=0 duration=2")

}
