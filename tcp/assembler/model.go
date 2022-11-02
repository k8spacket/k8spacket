package assembler

import (
	"github.com/google/gopacket"
	"github.com/k8spacket/k8s-api"
	"time"
)

type TcpStreamFactory struct{}

type tcpStream struct {
	streamId                                               uint32
	net, transport                                         gopacket.Flow
	bytesSent, bytesReceived, packets, outOfOrder, skipped int64
	start, end                                             time.Time
	sawStart, sawEnd                                       bool
}

var reverseLookupMap = make(map[string]string)

var K8sInfo = make(map[string]k8s.IPResourceInfo)
