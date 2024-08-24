package nodegraph_log

import "log"

var LOGGER log.Logger

func BuildLogger() {
	LOGGER = *log.New(log.Writer(), "[nodegraph module] ", log.LstdFlags|log.Lmsgprefix)
}
