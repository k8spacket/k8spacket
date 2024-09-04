package ebpf

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	ebpf_inet "github.com/k8spacket/k8spacket/ebpf/inet"
	ebpf_tc "github.com/k8spacket/k8spacket/ebpf/tc"
	"github.com/stretchr/testify/assert"
)

type mockInetEbpf struct {
	ebpf_inet.IInetEbpf
	initCalled bool
}

func (mockInetEbpf *mockInetEbpf) Init() {
	mockInetEbpf.initCalled = true
}

type mockItcEbpf struct {
	ebpf_tc.ItcEbpf
	initCalledCount int
}

func (mockItcEbpf *mockItcEbpf) Init(iface string) {
	mockItcEbpf.initCalledCount++
}

func TestLoad(t *testing.T) {

	var tests = []struct {
		command       string
		inetCalled    bool
		tcCalledCount int
		err           string
	}{
		{"echo 'iface1,iface2'", true, 2, ""},
		{"exit 1", true, 0, "[tc-loop] Cannot find interfaces to listen"},
	}

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	os.Setenv("K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD", "100ms")

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {

			os.Setenv("K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND", test.command)

			mockInetEbpf := &mockInetEbpf{}
			mockItcEbpf := &mockItcEbpf{}

			loader := Loader{inetEbpf: mockInetEbpf, tcEbpf: mockItcEbpf}

			loader.load()

			assert.Eventually(t, func() bool {
				return mockInetEbpf.initCalled == test.inetCalled && mockItcEbpf.initCalledCount == test.tcCalledCount && strings.Contains(str.String(), test.err)
			}, time.Second*1, time.Millisecond*100)
		})
	}
}
