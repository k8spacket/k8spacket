package backend

import (
	"encoding/json"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
)

type Handler struct {
	repo repository.Repository
}

func NewHandler(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (handler *Handler) TLSConnectionHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/tlsparser/connections/")
	if len(id) > 0 {
		w.Header().Set("Content-Type", "application/json")
		var details = handler.getConnection(id)
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
		err := json.NewEncoder(w).Encode(handler.filterConnections(req.URL.Query()))
		if err != nil {
			slog.Error("[api] Cannot prepare connections response", "Error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *Handler) getConnection(id string) model.TLSDetails {
	return handler.repo.Read(id)
}

func (handler *Handler) filterConnections(query url.Values) []model.TLSConnection {
	from := query["from"]
	rangeFrom := time.Time{}
	if len(from) > 0 {
		i, err := strconv.ParseInt(from[0], 10, 64)
		if err != nil {
			slog.Error("[api] cannot parse value", "Error", err)
		} else {
			rangeFrom = time.UnixMilli(i).UTC()
		}
	}

	to := query["to"]
	rangeTo := time.Time{}
	if len(to) > 0 {
		i, err := strconv.ParseInt(to[0], 10, 64)
		if err != nil {
			slog.Error("[api] cannot parse value", "Error", err)
		} else {
			rangeTo = time.UnixMilli(i).UTC()
		}
	}

	slog.Info("[api:params]", "from", rangeFrom, "to", rangeTo)
	return handler.repo.Query(rangeFrom, rangeTo)
}
