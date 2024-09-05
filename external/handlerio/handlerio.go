package handlerio

import "os"

type HandlerIO struct {
	IHandlerIO
}

func (handlerIO *HandlerIO) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
