package ebpf

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	ebpf_inet "github.com/k8spacket/k8spacket/internal/ebpf/inet"
	ebpf_socketfilter "github.com/k8spacket/k8spacket/internal/ebpf/socketfilter"
	ebpf_tc "github.com/k8spacket/k8spacket/internal/ebpf/tc"
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

type mockIsocketfilterEbpf struct {
	ebpf_socketfilter.ISocketFilterEbpf
	initCalledCount int
}

func (mockIsocketfilterEbpf *mockIsocketfilterEbpf) Init() {
	mockIsocketfilterEbpf.initCalledCount++
}

func TestLoad(t *testing.T) {

	var tests = []struct {
		command                 string
		loaderSource            string
		inetCalled              bool
		socketfilterCalledCount int
		tcCalledCount           int
		err                     string
	}{
		{"echo 'iface1,iface2'", "socketfilter", true, 1, 0, ""},
		{"echo 'iface1,iface2'", "tc", true, 0, 2, ""},
		{"echo 'iface1,iface2'", "", true, 1, 0, ""},
		{"echo 'iface1,iface2'", "some_other_value", true, 1, 0, ""},
		{"exit 1", "tc", true, 0, 0, "[tc-loop] Cannot find interfaces to listen"},
	}

	var str bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&str, nil))

	slog.SetDefault(logger)

	os.Setenv("K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD", "100ms")

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {

			os.Setenv("K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND", test.command)
			os.Setenv("K8S_PACKET_LOADER_SOURCE", test.loaderSource)

			mockInetEbpf := &mockInetEbpf{}
			mockItcEbpf := &mockItcEbpf{}
			mockIsocketfilterEbpf := &mockIsocketfilterEbpf{}
			loader := Init(mockInetEbpf, mockItcEbpf, mockIsocketfilterEbpf)
			loader.Load()

			assert.Eventually(t, func() bool {
				return mockInetEbpf.initCalled == test.inetCalled && mockItcEbpf.initCalledCount == test.tcCalledCount && mockIsocketfilterEbpf.initCalledCount == test.socketfilterCalledCount && strings.Contains(str.String(), test.err)
			}, time.Second*1, time.Millisecond*100)
		})
	}
}
