package ebpf_tools

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/k8spacket/k8spacket/internal/modules"
	"github.com/stretchr/testify/assert"
)

func TestEnrichAddress(t *testing.T) {

	os.Setenv("K8S_PACKET_REVERSE_WHOIS_REGEXP", "(?:OrgName:|org-name:)\\s*(.*)")
	os.Setenv("K8S_PACKET_REVERSE_GEOIP2_DB_PATH", "../../../tests/units/GeoLite2-City-Test.mmdb")

	address := modules.Address{Addr: "8.8.8.8"}

	EnrichAddress(&address)

	assert.EqualValues(t, "Google LLC", address.Name)

	address = modules.Address{Addr: "192.168.0.1"}

	EnrichAddress(&address)

	assert.EqualValues(t, "N/A", address.Name)

	address = modules.Address{Addr: "89.160.20.129", Port: 443}
	StoreDomain(address.Addr, address.Port, "89-160-20-129.cust.bredband2.com")

	EnrichAddress(&address)

	assert.EqualValues(t, "89-160-20-129.cust.bredband2.com, (SE, Linköping)", address.Name)

}

func TestHtonsBehavior(t *testing.T) {
	var a uint16 = 0x0102
	out := Htons(a)
	// expected swap
	expected := (a<<8)&0xff00 | (a>>8)&0x00ff
	assert.Equal(t, expected, out)
}

func TestIntToIP4(t *testing.T) {
	// bytes {1,2,3,4} as little endian uint32
	num := binary.LittleEndian.Uint32([]byte{1, 2, 3, 4})
	s := IntToIP4(num, binary.BigEndian.PutUint32)
	assert.Equal(t, "4.3.2.1", s)
	s = IntToIP4(num, binary.LittleEndian.PutUint32)
	assert.Equal(t, "1.2.3.4", s)
}
