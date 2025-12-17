package ebpf_tc

type Tc interface {
	Init(iface string)
}
