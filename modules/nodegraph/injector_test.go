package nodegraph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	listener := Init()

	assert.NotEmpty(t, listener)

}