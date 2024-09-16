generate:
	cd ./ebpf/inet
	go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target native -type event bpf ./bpf/inet.bpf.c
	cd ../../

	cd ./ebpf/tc
	go run github.com/cilium/ebpf/cmd/bpf2go tc ./bpf/tc.bpf.c
	cd ../../

fmt:
	go fmt ./...

build:
	go build .

test:
	K8S_PACKET_K8S_RESOURCES_DISABLED=true go test ./... -coverprofile=coverage.out

run:
	go run k8spacket.go

run_local:
	K8S_PACKET_TCP_LISTENER_PORT=6676 K8S_PACKET_TLS_CERTIFICATE_CACHE_TTL=30s K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND="echo -n eno2" K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD=3s K8S_PACKET_K8S_RESOURCES_DISABLED=true go run k8spacket.go
