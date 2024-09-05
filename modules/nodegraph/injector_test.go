package nodegraph

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	os.Setenv("K8S_PACKET_TCP_METRICS_ENABLED", "true")

	listener := Init(http.NewServeMux())

	assert.NotEmpty(t, listener)

}
