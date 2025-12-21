package backend

import (
	"encoding/json"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/repository"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
)

type Handler struct {
	repo repository.Repository[model.ConnectionItem]
}

func NewHandler(repo repository.Repository[model.ConnectionItem]) *Handler {
	return &Handler{repo: repo}
}

func (handler *Handler) ConnectionHandler(w http.ResponseWriter, r *http.Request) {
	var response = handler.filterConnections(r.URL.Query())

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("[api] Cannot prepare connections response", "Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *Handler) filterConnections(query url.Values) []model.ConnectionItem {
	var from = query["from"]
	var rangeFrom = time.Time{}
	if len(from) > 0 {
		i, err := strconv.ParseInt(from[0], 10, 64)
		if err != nil {
			slog.Error("[api] parse", "Error", err)
		}
		rangeFrom = time.UnixMilli(i)
	}

	var to = query["to"]
	var rangeTo = time.Time{}
	if len(to) > 0 {
		i, err := strconv.ParseInt(to[0], 10, 64)
		if err != nil {
			slog.Error("[api] parse", "Error", err)
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

	return handler.getConnections(rangeFrom, rangeTo, patternNs, patternIn, patternEx)
}

func (handler *Handler) getConnections(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []model.ConnectionItem {

	slog.Info("[api:params]",
		"patternNs", patternNs,
		"patternIn", patternIn,
		"patternEx", patternEx,
		"from", from.Format(time.DateTime),
		"to", to.Format(time.DateTime))

	return handler.repo.Query(from, to, patternNs, patternIn, patternEx)
}
