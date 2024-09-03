package tlsparser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	os.Setenv("K8S_PACKET_TLS_METRICS_ENABLED", "true")

	listener := Init()

	assert.NotEmpty(t, listener)

}
