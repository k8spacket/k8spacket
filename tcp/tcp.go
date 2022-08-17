package tcp

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/reassembly"
	"github.com/k8spacket/k8s"
	"github.com/k8spacket/tcp/assembler"
	"github.com/k8spacket/tools"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func listenInterface(iface string, filter string) {

	log.Printf("Starting capture on interface %q", iface)
	handle, err := pcap.OpenLive(iface, 262144, true, pcap.BlockForever)
	if err != nil {
		log.Fatal("error opening pcap handle: ", err)
	}
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatal("error setting BPF filter: ", err)
	}

	streamFactory := &assembler.TcpStreamFactory{}
	streamPool := reassembly.NewStreamPool(streamFactory)
	tcpassembler := reassembly.NewAssembler(streamPool)
	var maxPagesPerConn, _ = strconv.Atoi(os.Getenv("K8S_PACKET_TCP_ASSEMBLER_MAX_PAGES_PER_CONN"))
	tcpassembler.MaxBufferedPagesPerConnection = maxPagesPerConn
	var maxPagesTotal, _ = strconv.Atoi(os.Getenv("K8S_PACKET_TCP_ASSEMBLER_MAX_PAGES_TOTAL"))
	tcpassembler.MaxBufferedPagesTotal = maxPagesTotal

	log.Println("reading in packets")

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	var period, _ = time.ParseDuration(os.Getenv("K8S_PACKET_TCP_ASSEMBLER_FLUSHING_PERIOD"))
	var closeOlderThan, _ = time.ParseDuration(os.Getenv("K8S_PACKET_TCP_ASSEMBLER_FLUSHING_CLOSE_OLDER_THAN"))
	ticker := time.Tick(period)
	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			tcpassembler.Assemble(packet.NetworkLayer().NetworkFlow(), tcp)
		case <-ticker:
			// Every (period) seconds, flush connections that haven't seen activity in the past (closeOlderThan) seconds.
			tcpassembler.FlushCloseOlderThan(time.Now().Add(-closeOlderThan))
			if !tools.SliceContains(interfaces, iface) {
				log.Printf("Stopping capture on interface %q", iface)
				return
			}
		}
	}
}

func findInterfaces() []string {
	command := os.Getenv("K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND")
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return nil
	}
	return strings.Split(string(out), ",")
}

var interfaces []string

func interfacesRefresher() {
	var currentInterfaces []string
	for {
		log.Printf("Refreshing interfaces for capturing...")
		interfaces = findInterfaces()
		var refreshK8sInfo = false
		for _, el := range interfaces {
			if (strings.TrimSpace(el) != "") && (!tools.SliceContains(currentInterfaces, el)) {
				go listenInterface(el, "tcp")
				refreshK8sInfo = true
			}
		}
		if refreshK8sInfo {
			k8s.FetchK8SInfo()
		}
		currentInterfaces = interfaces
		var refreshPeriod, _ = time.ParseDuration(os.Getenv("K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD"))
		time.Sleep(refreshPeriod)
	}
}

func StartListeners() {
	defer util.Run()()
	go interfacesRefresher()

}
