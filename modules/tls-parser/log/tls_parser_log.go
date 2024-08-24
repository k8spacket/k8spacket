package tls_parser_log

import "log"

var LOGGER log.Logger

func BuildLogger() {
	LOGGER = *log.New(log.Writer(), "[tls-parser module] ", log.LstdFlags|log.Lmsgprefix)
}
