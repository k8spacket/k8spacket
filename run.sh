#!/usr/bin/env bash

#set -x

# pushd ebpf/inet
# pushd bpf
# ../../../libbpf.sh
# popd
# go generate -ldflags "-w -s"
# popd

# pushd ebpf/tc
# pushd bpf
# ../../../libbpf.sh
# popd
# go generate -ldflags "-w -s"
# popd

go build .

K8S_PACKET_TCP_LISTENER_PORT=6676 K8S_PACKET_TLS_CERTIFICATE_CACHE_TTL=30s K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND="echo -n eno2" K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD=3s K8S_PACKET_K8S_RESOURCES_DISABLED=true go run k8spacket.go

echo "Stop"
