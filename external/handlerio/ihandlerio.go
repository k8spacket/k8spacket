package handlerio

type IHandlerIO interface {
	ReadFile(name string) ([]byte, error)
}
