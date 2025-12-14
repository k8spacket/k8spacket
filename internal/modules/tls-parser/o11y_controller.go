package tlsparser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/k8spacket/k8spacket/internal/modules/tls-parser/model"
)

const connectionDetailsUri = "/tlsparser/api/data/"

type O11yController struct {
	service IService
}

func (o11yController *O11yController) TLSParserConnectionsHandler(w http.ResponseWriter, req *http.Request) {
	out, err := o11yController.service.buildConnectionsResponse(fmt.Sprintf("http://%%s:%s/tlsparser/connections/?%s", os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), req.URL.Query().Encode()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prepareResponse(w, out)
}

func (o11yController *O11yController) TLSParserConnectionDetailsHandler(w http.ResponseWriter, req *http.Request) {
	idParam := strings.TrimPrefix(req.URL.Path, connectionDetailsUri)
	if len(strings.TrimSpace(idParam)) > 0 {
		out, err := o11yController.service.buildDetailsResponse(fmt.Sprintf("http://%%s:%s/tlsparser/connections/%s?%s", os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), idParam, req.URL.Query().Encode()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prepareResponse(w, out)
	} else {
		o11yController.TLSParserConnectionsHandler(w, req)
	}
}

func prepareResponse[T model.TLSDetails | []model.TLSConnection](w http.ResponseWriter, out T) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		slog.Error("[api] Cannot prepare stats response", "Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
