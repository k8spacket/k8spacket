package k8spacket_log

import "log"

var (
	LOGGER log.Logger
)

func BuildLogger() {
	LOGGER = *log.New(log.Writer(), "[k8spacket] ", log.LstdFlags|log.Lmsgprefix)
}
