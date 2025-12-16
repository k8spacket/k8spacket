package io

type IO interface {
	Read(name string) ([]byte, error)
}
