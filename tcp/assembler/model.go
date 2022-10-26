package assembler

import (
	"github.com/google/gopacket"
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
