generate:
	pushd ebpf/inet
	go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target native -type event bpf ./bpf/inet.bpf.c
	popd

	pushd ebpf/tc
	go run github.com/cilium/ebpf/cmd/bpf2go tc ./bpf/tc.bpf.c
	popd

build:
	go build .

run:
	go run k8spacket.go