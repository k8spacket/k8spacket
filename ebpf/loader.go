package ebpf

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	ebpf_inet "github.com/k8spacket/k8spacket/ebpf/inet"
	ebpf_tc "github.com/k8spacket/k8spacket/ebpf/tc"
)

type Loader struct {
	inetEbpf   ebpf_inet.IInetEbpf
	tcEbpf     ebpf_tc.ItcEbpf
	interfaces []string
}

func Init(inetEbpf ebpf_inet.IInetEbpf, tcEbpf ebpf_tc.ItcEbpf) *Loader {
	return &Loader{inetEbpf: inetEbpf, tcEbpf: tcEbpf}
}

func (loader *Loader) Load() {
	// load inet_sock_set_state ebpf program
	go loader.inetEbpf.Init()
	go interfacesRefresher(*loader)
}

func interfacesRefresher(loader Loader) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var currentInterfaces []string
	var refreshPeriod, _ = time.ParseDuration(os.Getenv("K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD"))

	for {
		select {
		case <-ctx.Done():
			slog.Info("[tc-loop] Receive signal, exiting...")
			return
		case <-time.After(refreshPeriod):
			slog.Info("[tc-loop] Refreshing interfaces for capturing...")
			loader.interfaces = findInterfaces()
			for _, el := range loader.interfaces {
				if (strings.TrimSpace(el) != "") && (!slices.Contains(currentInterfaces, el)) {
					// load traffic control ebpf program (qdisc filter)
					go loader.tcEbpf.Init(el)
				}
			}
			currentInterfaces = loader.interfaces
		}
	}
}

// looking for network interfaces on cluster nodes regarding started containers based on the command `ip address`
func findInterfaces() []string {
	command := os.Getenv("K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND")
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.Output()

	if err != nil {
		slog.Error("[tc-loop] Cannot find interfaces to listen", "Error", err)
		return nil
	}
	return strings.Split(string(out), ",")
}
