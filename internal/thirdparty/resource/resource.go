package resource

type Resource interface {
	Read(name string) ([]byte, error)
}
