package assembler

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"github.com/k8spacket/k8s-api"
	"github.com/k8spacket/k8spacket/broker"
	"github.com/k8spacket/plugin-api"
	"github.com/likexian/whois"
	"github.com/oschwald/geoip2-golang"
	"math/rand"
	"net"
	"os"
	"regexp"
	"time"
)

func (s *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	payload := plugin_api.TCPPacketPayload{StreamId: s.streamId, Payload: tcp.Payload}
	broker.TCPPacketPayoutEvent(payload)
	return true
}

func (factory *TcpStreamFactory) New(netFlow, tcpFlow gopacket.Flow, _ *layers.TCP, _ reassembly.AssemblerContext) reassembly.Stream {
	return &tcpStream{
		streamId:  rand.Uint32(),
		net:       netFlow,
		transport: tcpFlow,
		start:     time.Now(),
	}
}

func (s *tcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	direction, start, end, skip := sg.Info()
	if start {
		s.start = ac.GetCaptureInfo().Timestamp
		s.end = s.start
	}
	if end {
		s.end = ac.GetCaptureInfo().Timestamp
	}
	if ac.GetCaptureInfo().Timestamp.Before(s.end) {
		s.outOfOrder++
	}
	aval, _ := sg.Lengths()
	if !direction {
		s.bytesSent += int64(aval)
	} else {
		s.bytesReceived += int64(aval)
	}
	s.packets += 1
	if skip > 0 {
		s.skipped += int64(skip)
	}
	s.sawStart = s.sawStart || start
	s.sawEnd = s.sawEnd || end
}

func (s *tcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	if s.sawStart {
		stream := plugin_api.ReassembledStream{
			StreamId:      s.streamId,
			Src:           s.net.Src().String(),
			SrcPort:       s.transport.Src().String(),
			Dst:           s.net.Dst().String(),
			DstPort:       s.transport.Dst().String(),
			Closed:        s.sawEnd,
			BytesSent:     float64(s.bytesSent),
			BytesReceived: float64(s.bytesReceived),
			Duration:      s.end.Sub(s.start).Seconds()}
		enrichStream(&stream)
		broker.ReassembledStreamEvent(stream)
	}
	return true
}

func enrichStream(stream *plugin_api.ReassembledStream) {
	var srcName = k8s.K8sInfo[stream.Src].Name
	if srcName == "" {
		stream.SrcName = reverseLookup(stream.Src)
	}

	var dstName = k8s.K8sInfo[stream.Dst].Name
	if dstName == "" {
		stream.DstName = reverseLookup(stream.Dst)
	}
	stream.SrcNamespace = k8s.K8sInfo[stream.Src].Namespace
	stream.DstNamespace = k8s.K8sInfo[stream.Dst].Namespace
}

func reverseLookup(ip string) string {

	if privateIPCheck(ip) {
		return "N/A"
	}

	if _, ok := reverseLookupMap[ip]; !ok {

		result, _ := whois.Whois(ip)

		re := regexp.MustCompile(os.Getenv("K8S_PACKET_REVERSE_WHOIS_REGEXP"))
		matches := re.FindStringSubmatch(result)

		reverseLookup := ""

		if len(matches) > 1 {
			reverseLookup += matches[1]
		}

		db, err := geoip2.Open(os.Getenv("K8S_PACKET_REVERSE_GEOIP2_DB_PATH"))
		if err == nil {
			defer db.Close()

			ipObj := net.ParseIP(ip)
			record, _ := db.City(ipObj)
			reverseLookup += "(" + record.Country.IsoCode + ", " + record.City.Names["en"] + ")"
		}
		reverseLookupMap[ip] = reverseLookup
	}
	return reverseLookupMap[ip]
}

// Check if an ip is private.
func privateIPCheck(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress.IsPrivate()
}
