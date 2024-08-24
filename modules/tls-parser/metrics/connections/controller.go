package connections

import (
	"encoding/json"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	tls_connection_db "github.com/k8spacket/k8spacket/modules/tls-parser/metrics/db/tls_connection"
	tls_detail_db "github.com/k8spacket/k8spacket/modules/tls-parser/metrics/db/tls_detail"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func TLSConnectionHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/tlsparser/connections/")
	if len(id) > 0 {
		w.Header().Set("Content-Type", "application/json")
		var details = tls_detail_db.Read(id)
		if !reflect.DeepEqual(details, model.TLSDetails{}) {
			err := json.NewEncoder(w).Encode(details)
			if err != nil {
				tls_parser_log.LOGGER.Printf("[api] Cannot prepare connection details response: %+v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found 404"))
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(filterConnections(req.URL.Query()))
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot prepare connections response: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func filterConnections(query url.Values) []model.TLSConnection {
	var from = query["from"]
	var rangeFrom = time.Time{}
	if len(from) > 0 {
		i, err := strconv.ParseInt(from[0], 10, 64)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] parse: %+v", err)
		}
		rangeFrom = time.UnixMilli(i)
	}

	var to = query["to"]
	var rangeTo = time.Time{}
	if len(to) > 0 {
		i, err := strconv.ParseInt(to[0], 10, 64)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] parse: %+v", err)
		}
		rangeTo = time.UnixMilli(i)
	}

	tls_parser_log.LOGGER.Printf("[api:params] from: %s, to: %s", rangeFrom, rangeTo)
	return tls_connection_db.Query(rangeFrom, rangeTo)
}
