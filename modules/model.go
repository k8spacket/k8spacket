package modules

type Address struct {
	Addr      string
	Port      uint16
	Name      string
	Namespace string
}
type TCPEvent struct {
	Client  Address
	Server  Address
	TxB     uint64
	RxB     uint64
	DeltaUs uint64
}

type TLSEvent struct {
	Client         Address
	Server         Address
	TlsVersions    []uint16
	Ciphers        []uint16
	ServerName     string
	UsedTlsVersion uint16
	UsedCipher     uint16
}
