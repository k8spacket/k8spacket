package resource

import "os"

type FileResource struct {
	Resource
}

func (fileResource *FileResource) Read(name string) ([]byte, error) {
	return os.ReadFile(name)
}
