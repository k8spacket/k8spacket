package ebpf

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/k8spacket/k8s-api/v2"
	"github.com/k8spacket/k8spacket/broker"
	ebpf_inet "github.com/k8spacket/k8spacket/ebpf/inet"
	ebpf_tc "github.com/k8spacket/k8spacket/ebpf/tc"
	ebpf_tools "github.com/k8spacket/k8spacket/ebpf/tools"
)

func LoadEbpf(broker broker.IBroker) {
	// load inet_sock_set_state ebpf program
	inetEbpf := ebpf_inet.InetEbpf{Broker: broker}
	tcEbpf := ebpf_tc.TcEbpf{Broker: broker}
	go inetEbpf.Init()
	go interfacesRefresher(tcEbpf)
}

func interfacesRefresher(tc ebpf_tc.TcEbpf) {
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
			interfaces = findInterfaces()
			var refreshK8sInfo = false
			for _, el := range interfaces {
				if (strings.TrimSpace(el) != "") && (!ebpf_tools.SliceContains(currentInterfaces, el)) {
					// load traffic control ebpf program (qdisc filter)
					go tc.Init(el)
					refreshK8sInfo = true
				}
			}
			if refreshK8sInfo {
				// there are some new workloads in the cluster and need to update info about k8s resources
				ebpf_tools.K8sInfo = k8s.FetchK8SInfo()
			}
			currentInterfaces = interfaces
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

var interfaces []string
