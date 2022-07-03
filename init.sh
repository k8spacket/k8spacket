#bash

export GO111MODULE=on
export CGO_ENABLED=1
go mod init github.com/k8spacket
go get github.com/google/gopacket
go get github.com/imdario/mergo
go get github.com/inhies/go-bytesize
go get github.com/likexian/whois
go get github.com/oschwald/geoip2-golang
go mod tidy
go build .
