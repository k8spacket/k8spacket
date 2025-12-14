package modules

type IListener[T TCPEvent | TLSEvent] interface {
	Listen(event T)
}
