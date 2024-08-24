package nodegraph

import (
	"encoding/json"
	nodegraph_log "github.com/k8spacket/k8spacket/modules/nodegraph/log"
	tcp_connection_db "github.com/k8spacket/k8spacket/modules/nodegraph/metrics/db/tcp_connection"
	"github.com/k8spacket/k8spacket/modules/nodegraph/metrics/nodegraph/model"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func ConnectionHandler(w http.ResponseWriter, r *http.Request) {
	connectionItemsMutex.RLock()
	var response = filterConnections(r.URL.Query())
	connectionItemsMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		nodegraph_log.LOGGER.Printf("[api] Cannot prepare connections response: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func filterConnections(query url.Values) []model.ConnectionItem {
	var from = query["from"]
	var rangeFrom = time.Time{}
	if len(from) > 0 {
		i, err := strconv.ParseInt(from[0], 10, 64)
		if err != nil {
			nodegraph_log.LOGGER.Printf("[api] parse: %+v", err)
		}
		rangeFrom = time.UnixMilli(i)
	}

	var to = query["to"]
	var rangeTo = time.Time{}
	if len(to) > 0 {
		i, err := strconv.ParseInt(to[0], 10, 64)
		if err != nil {
			nodegraph_log.LOGGER.Printf("[api] parse: %+v", err)
		}
		rangeTo = time.UnixMilli(i)
	}

	var namespace = query["namespace"]
	var patternNs = regexp.MustCompile("")
	if len(namespace) > 0 {
		patternNs = regexp.MustCompile(namespace[0])
	}

	var include = query["include"]
	var patternIn = regexp.MustCompile("")
	if len(include) > 0 {
		patternIn = regexp.MustCompile(include[0])
	}

	var exclude = query["exclude"]
	var patternEx = regexp.MustCompile("")
	if len(exclude) > 0 {
		patternEx = regexp.MustCompile(exclude[0])
	}

	nodegraph_log.LOGGER.Printf("[api:params] patternNs: %s, patternIn: %s, patternEx: %s, from: %s, to: %s", patternNs, patternIn, patternEx, rangeFrom, rangeTo)
	return tcp_connection_db.Query(rangeFrom, rangeTo, patternNs, patternIn, patternEx)
}
