#bash

export GO111MODULE=on
export CGO_ENABLED=1
go mod init github.com/k8spacket
go build .
