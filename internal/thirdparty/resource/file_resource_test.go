package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileResource_Read_Success(t *testing.T) {
	base := "../../../tests/units"
	fname := base + "/file_resource.txt"

	fr := &FileResource{}
	got, err := fr.Read(fname)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello k8spacket\n"), got)
}

func TestFileResource_Read_NotFound(t *testing.T) {
	base := "../../../tests/units"
	missing := base + "/does-not-exist.txt"

	fr := &FileResource{}
	_, err := fr.Read(missing)
	assert.Error(t, err)
}
