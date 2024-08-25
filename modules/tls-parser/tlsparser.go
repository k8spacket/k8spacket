package tlsparser

import (
	"github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/connections"
	"net/http"
)

func init() {
	tls_parser_log.BuildLogger()

	http.HandleFunc("/tlsparser/connections/", connections.TLSConnectionHandler)
	http.HandleFunc("/tlsparser/api/data", connections.TLSParserConnectionsHandler)
	http.HandleFunc("/tlsparser/api/data/", connections.TLSParserConnectionDetailsHandler)
}
