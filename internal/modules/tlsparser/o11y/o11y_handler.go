package o11y

import (
	"encoding/json"
	"fmt"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
	k8sclient "github.com/k8spacket/k8spacket/internal/thirdparty/k8s"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
)

const connectionDetailsUri = "/tlsparser/api/data/"

type O11yHandler struct {
	httpClient httpclient.Client
	k8sClient  k8sclient.Client
}

func NewO11yHandler(httpClient httpclient.Client, k8sClient k8sclient.Client) *O11yHandler {
	return &O11yHandler{httpClient: httpClient, k8sClient: k8sClient}
}

func (handler *O11yHandler) TLSParserConnectionsHandler(w http.ResponseWriter, req *http.Request) {
	out, err := handler.buildConnectionsResponse(fmt.Sprintf("http://%%s:%s/tlsparser/connections/?%s", os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), req.URL.Query().Encode()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prepareResponse(w, out)
}

func (handler *O11yHandler) TLSParserConnectionDetailsHandler(w http.ResponseWriter, req *http.Request) {
	idParam := strings.TrimPrefix(req.URL.Path, connectionDetailsUri)
	if len(strings.TrimSpace(idParam)) > 0 {
		out, err := handler.buildDetailsResponse(fmt.Sprintf("http://%%s:%s/tlsparser/connections/%s?%s", os.Getenv("K8S_PACKET_TCP_LISTENER_PORT"), idParam, req.URL.Query().Encode()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prepareResponse(w, out)
	} else {
		handler.TLSParserConnectionsHandler(w, req)
	}
}

func (handler *O11yHandler) buildConnectionsResponse(url string) ([]model.TLSConnection, error) {
	resultFunc := func(destination, source []model.TLSConnection) []model.TLSConnection {
		return append(destination, source...)
	}
	return buildResponse(handler, url, []model.TLSConnection{}, resultFunc)
}

func (handler *O11yHandler) buildDetailsResponse(url string) (model.TLSDetails, error) {
	resultFunc := func(destination, source model.TLSDetails) model.TLSDetails {
		if !reflect.DeepEqual(source, model.TLSDetails{}) {
			return source
		} else {
			return destination
		}
	}
	return buildResponse(handler, url, model.TLSDetails{}, resultFunc)
}

func buildResponse[T model.TLSDetails | []model.TLSConnection](handler *O11yHandler, url string, t T, resultFunc func(d T, s T) T) (T, error) {
	var k8spacketIps = handler.k8sClient.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))

	var in T
	out := t

	for _, ip := range k8spacketIps {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(url, ip), nil)
		resp, err := handler.httpClient.Do(req)

		if err != nil {
			slog.Error("[api] Cannot get stats", "Error", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {

			responseData, err := io.ReadAll(resp.Body)
			if err != nil {
				slog.Error("[api] Cannot read stats response", "Error", err)
				continue
			}

			err = json.Unmarshal(responseData, &in)
			if err != nil {
				slog.Error("[api] Cannot parse stats response", "Error", err)
				continue
			}

			out = resultFunc(out, in)
		}
	}

	return out, nil
}

func prepareResponse[T model.TLSDetails | []model.TLSConnection](w http.ResponseWriter, out T) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		slog.Error("[api] Cannot prepare stats response", "Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
