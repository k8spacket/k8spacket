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

	"github.com/k8spacket/k8spacket/broker"
	ebpf_inet "github.com/k8spacket/k8spacket/ebpf/inet"
	ebpf_tc "github.com/k8spacket/k8spacket/ebpf/tc"
	ebpf_tools "github.com/k8spacket/k8spacket/ebpf/tools"
	k8sclient "github.com/k8spacket/k8spacket/external/k8s"
)

type Loader struct {
	inetEbpf   ebpf_inet.IInetEbpf
	tcEbpf     ebpf_tc.ItcEbpf
	interfaces []string
}

func Init(broker broker.IBroker) {

	inetEbpf := &ebpf_inet.InetEbpf{Broker: broker}
	tcEbpf := &ebpf_tc.TcEbpf{Broker: broker}

	loader := Loader{inetEbpf: inetEbpf, tcEbpf: tcEbpf}

	loader.load()
}

func (loader *Loader) load() {
	// load inet_sock_set_state ebpf program
	go loader.inetEbpf.Init()
	go loader.interfacesRefresher(loader.tcEbpf)
}

func (loader *Loader) interfacesRefresher(tc ebpf_tc.ItcEbpf) {
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
			var refreshK8sInfo = false
			for _, el := range loader.interfaces {
				if (strings.TrimSpace(el) != "") && (!ebpf_tools.SliceContains(currentInterfaces, el)) {
					// load traffic control ebpf program (qdisc filter)
					go tc.Init(el)
					refreshK8sInfo = true
				}
			}
			if refreshK8sInfo {
				// there are some new workloads in the cluster and need to update info about k8s resources
				ebpf_tools.K8sInfo = k8sclient.FetchK8SInfo()
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
