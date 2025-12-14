package ebpf_tools

import (
	"os"
	"testing"

	"github.com/k8spacket/k8spacket/pkg/events"
	"github.com/stretchr/testify/assert"
)

func TestEnrichAddress(t *testing.T) {

	os.Setenv("K8S_PACKET_REVERSE_WHOIS_REGEXP", "(?:OrgName:|org-name:)\\s*(.*)")
	os.Setenv("K8S_PACKET_REVERSE_GEOIP2_DB_PATH", "../../../tests/units/GeoLite2-City-Test.mmdb")

	address := events.Address{Addr: "8.8.8.8"}

	EnrichAddress(&address)

	assert.EqualValues(t, "Google LLC", address.Name)

	address = events.Address{Addr: "192.168.0.1"}

	EnrichAddress(&address)

	assert.EqualValues(t, "N/A", address.Name)

	address = events.Address{Addr: "89.160.20.129", Port: 443}
	StoreDomain(address.Addr, address.Port, "89-160-20-129.cust.bredband2.com")

	EnrichAddress(&address)

	assert.EqualValues(t, "89-160-20-129.cust.bredband2.com, (SE, Linköping)", address.Name)

}
