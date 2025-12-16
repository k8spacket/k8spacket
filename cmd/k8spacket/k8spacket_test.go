package main

import (
	"github.com/k8spacket/k8spacket/internal/ebpf"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/broker"
	"github.com/stretchr/testify/assert"
)

type mockLoader struct {
	ebpf.Loader
}

func (mockLoader *mockLoader) Load() {

}

func TestStartApp(t *testing.T) {

	os.Setenv("K8S_PACKET_TCP_LISTENER_PORT", "6676")

	mux := http.NewServeMux()

	b := &broker.DistributionBroker{}
	loader := &mockLoader{}

	go startApp(b, loader, mux)

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://127.0.0.1:6676/metrics")
		if err != nil {
			return false
		}
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		return assert.EqualValues(t, resp.StatusCode, http.StatusOK) && assert.Regexp(t, "go_info{version=\"go.*\"}", bodyStr)
	}, time.Second*20, time.Millisecond*100)

}
