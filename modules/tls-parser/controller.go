package tlsparser

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
)

type Controller struct {
	service IService
}

func (controller *Controller) TLSConnectionHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/tlsparser/connections/")
	if len(id) > 0 {
		w.Header().Set("Content-Type", "application/json")
		var details = controller.service.getConnection(id)
		if !reflect.DeepEqual(details, model.TLSDetails{}) {
			err := json.NewEncoder(w).Encode(details)
			if err != nil {
				slog.Error("[api] Cannot prepare connection details response", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found 404"))
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(controller.service.filterConnections(req.URL.Query()))
		if err != nil {
			slog.Error("[api] Cannot prepare connections response", "Error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
