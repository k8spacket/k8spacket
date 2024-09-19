.ONESHELL:
prepare:
	cd ./ebpf/inet/bpf
	./../../../libbpf.sh
	cd ../../../

	cd ./ebpf/tc/bpf
	./../../../libbpf.sh
	cd ../../../

.ONESHELL:
generate: prepare
	cd ./ebpf/inet
	go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target native -type event -go-package ebpf_inet bpf ./bpf/inet.bpf.c
	cd ../../

	cd ./ebpf/tc
	go run github.com/cilium/ebpf/cmd/bpf2go -go-package ebpf_tc tc ./bpf/tc.bpf.c
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

.ONESHELL:
start_qemu:
	cd ${GITHUB_WORKSPACE}/tests/e2e/vm/filesystem
	unzip ./filesystem.zip
	sudo qemu-img create -f qcow2 -b filesystem.qcow2 -F qcow2 filesystem-diff.qcow2
	sudo qemu-system-x86_64 \
	-cpu host \
	-m 4G \
	-smp 4 \
	-kernel ${GITHUB_WORKSPACE}/tests/e2e/vm/kernels/${KERNEL}/bzImage \
	-append "console=ttyS0 root=/dev/sda rw" \
	-drive file=${GITHUB_WORKSPACE}/tests/e2e/vm/filesystem/filesystem-diff.qcow2,format=qcow2 \
	-net nic -net user,hostfwd=tcp::10022-:22,hostfwd=tcp::16676-:6676,hostfwd=tcp::10443-:443 \
	-enable-kvm \
	-nographic &

.ONESHELL:
prepare_e2e: start_qemu
	cd ${GITHUB_WORKSPACE}
	while ! nc -z 127.0.0.1 10022 ; do echo "waiting for ssh"; sleep 1; done
	sshpass -p root scp -o 'StrictHostKeyChecking no' -P 10022 ./k8spacket root@127.0.0.1:/root/k8spacket
	sshpass -p root scp -o 'StrictHostKeyChecking no' -P 10022 ./fields.json root@127.0.0.1:/root/fields.json
	sshpass -p root ssh -p 10022 root@127.0.0.1 'chmod 0655 /root/k8spacket && systemctl start k8spacket.service'
	while ! sshpass -p root ssh -p 10022 root@127.0.0.1 'systemctl is-active k8spacket.service' ; do echo "waiting for k8spacket service"; sleep 1; done

.ONESHELL:
e2e: prepare_e2e
	cd ${GITHUB_WORKSPACE}/tests/e2e
	go test