package api

type IListener[T interface{}] interface {
	Listen(event T)
}
