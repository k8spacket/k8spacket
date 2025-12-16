package modules

type Listener[T TCPEvent | TLSEvent] interface {
	Listen(event T)
}
