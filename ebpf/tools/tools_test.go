package ebpf_tools

import (
	"os"
	"testing"

	"github.com/k8spacket/k8spacket/modules"
	"github.com/stretchr/testify/assert"
)

func TestEnrichAddress(t *testing.T) {

	os.Setenv("K8S_PACKET_REVERSE_WHOIS_REGEXP", "(?:OrgName:|org-name:)\\s*(.*)")
	os.Setenv("K8S_PACKET_REVERSE_GEOIP2_DB_PATH", "../../tests/GeoLite2-City-Test.mmdb")

	address := modules.Address{Addr: "89.160.20.129"}

	EnrichAddress(&address)

	assert.EqualValues(t, "(SE, Linköping)", address.Name)

	address = modules.Address{Addr: "8.8.8.8"}

	EnrichAddress(&address)

	assert.EqualValues(t, "Google LLC", address.Name)

	address = modules.Address{Addr: "192.168.0.1"}

	EnrichAddress(&address)

	assert.EqualValues(t, "N/A", address.Name)

}

func TestSliceContains(t *testing.T) {

	slice := []string{"A", "B", "C"}

	assert.EqualValues(t, true, SliceContains(slice, "B"))
	assert.EqualValues(t, false, SliceContains(slice, "D"))

}
