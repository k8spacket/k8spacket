package io

import "os"

type FileIO struct {
	IO
}

func (fileIO *FileIO) Read(name string) ([]byte, error) {
	return os.ReadFile(name)
}
