package assembler

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"github.com/k8spacket/metrics"
	"time"
)

type TcpStreamFactory struct{}

type tcpStream struct {
	net, transport                                         gopacket.Flow
	bytesSent, bytesReceived, packets, outOfOrder, skipped int64
	start, end                                             time.Time
	sawStart, sawEnd                                       bool
}

func (s *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	return true
}

func (factory *TcpStreamFactory) New(netFlow, tcpFlow gopacket.Flow, _ *layers.TCP, _ reassembly.AssemblerContext) reassembly.Stream {
	return &tcpStream{
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
		metrics.PushK8sPacketMetric(s.net.Src().String(), s.transport.Src().String(), s.net.Dst().String(), s.transport.Dst().String(), s.sawEnd, float64(s.bytesSent), float64(s.bytesReceived), s.end.Sub(s.start).Seconds())
	}
	return true
}
