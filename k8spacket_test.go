package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/broker"
	ebpf_inet "github.com/k8spacket/k8spacket/ebpf/inet"
	ebpf_tc "github.com/k8spacket/k8spacket/ebpf/tc"
	"github.com/stretchr/testify/assert"
)

type mockLoader struct {
	inetEbpf ebpf_inet.IInetEbpf
	tcEbpf   ebpf_tc.ItcEbpf
}

func (mockLoader *mockLoader) Load() {

}

func TestStartApp(t *testing.T) {

	os.Setenv("K8S_PACKET_TCP_LISTENER_PORT", "6676")

	mux := http.NewServeMux()

	broker := &broker.Broker{}
	loader := &mockLoader{}

	go startApp(broker, loader, mux)

	assert.Eventually(t, func() bool {
		resp, _ := http.Get("http://127.0.0.1:6676/metrics")
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		return resp.StatusCode == http.StatusOK && strings.Contains(bodyStr, "go_info{version=\"go1.23.0\"}")
	}, time.Second*2, time.Millisecond*100)

}
