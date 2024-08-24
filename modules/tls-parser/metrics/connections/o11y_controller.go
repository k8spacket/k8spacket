package connections

import (
	"encoding/json"
	"fmt"
	"github.com/k8spacket/k8s-api/v2"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/metrics/model"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
)

const connectionDetailsUri = "/tlsparser/api/data/"

func TLSParserConnectionsHandler(w http.ResponseWriter, req *http.Request) {
	resultFunc := func(destination, source []model.TLSConnection) []model.TLSConnection {
		return append(destination, source...)
	}
	buildResponse(w, fmt.Sprintf("http://%%s:%s/tlsparser/connections/?%s", os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), req.URL.Query().Encode()), []model.TLSConnection{}, resultFunc)
}

func TLSParserConnectionDetailsHandler(w http.ResponseWriter, req *http.Request) {
	idParam := strings.TrimPrefix(req.URL.Path, connectionDetailsUri)
	if len(strings.TrimSpace(idParam)) > 0 {
		resultFunc := func(destination, source model.TLSDetails) model.TLSDetails {
			if !reflect.DeepEqual(source, model.TLSDetails{}) {
				return source
			} else {
				return destination
			}
		}
		buildResponse(w, fmt.Sprintf("http://%%s:%s/tlsparser/connections/%s?%s", os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), idParam, req.URL.Query().Encode()), model.TLSDetails{}, resultFunc)
	} else {
		TLSParserConnectionsHandler(w, req)
	}
}

func buildResponse[T model.TLSDetails | []model.TLSConnection](w http.ResponseWriter, url string, t T, resultFunc func(d T, s T) T) {
	var k8spacketIps = k8s.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))

	var in T
	out := t

	for _, ip := range k8spacketIps {
		resp, err := http.Get(fmt.Sprintf(url, ip))

		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot get stats: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot read stats response: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		_ = json.Unmarshal(responseData, &in)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot parse stats response: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		out = resultFunc(out, in)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		tls_parser_log.LOGGER.Printf("[api] Cannot prepare stats response: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
