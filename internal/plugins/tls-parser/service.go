package tlsparser

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/k8spacket/k8spacket/internal/infra/db"
	httpclient "github.com/k8spacket/k8spacket/internal/infra/http"
	"github.com/k8spacket/k8spacket/internal/infra/k8s"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/certificate"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/model"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/repository"
)

type Service struct {
	repo        repository.IRepository
	certificate certificate.ICertificate
	httpClient  httpclient.IHttpClient
	k8sClient   k8sclient.IK8SClient
}

func (service *Service) storeInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails) {
	var id = strconv.Itoa(int(db.HashId(fmt.Sprintf("%s-%s", tlsConnection.Src, tlsConnection.Dst))))
	tlsConnection.Id = id
	service.repo.UpsertConnection(id, tlsConnection)
	tlsDetails.Id = id
	service.repo.UpsertDetails(id, tlsDetails, service.certificate.UpdateCertificateInfo)
}

func (service *Service) getConnection(id string) model.TLSDetails {
	return service.repo.Read(id)
}

func (service *Service) filterConnections(query url.Values) []model.TLSConnection {
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
	return service.repo.Query(rangeFrom, rangeTo)
}

func (service *Service) buildConnectionsResponse(url string) ([]model.TLSConnection, error) {
	resultFunc := func(destination, source []model.TLSConnection) []model.TLSConnection {
		return append(destination, source...)
	}
	return buildResponse(service, url, []model.TLSConnection{}, resultFunc)
}

func (service *Service) buildDetailsResponse(url string) (model.TLSDetails, error) {
	resultFunc := func(destination, source model.TLSDetails) model.TLSDetails {
		if !reflect.DeepEqual(source, model.TLSDetails{}) {
			return source
		} else {
			return destination
		}
	}
	return buildResponse(service, url, model.TLSDetails{}, resultFunc)
}

func buildResponse[T model.TLSDetails | []model.TLSConnection](service *Service, url string, t T, resultFunc func(d T, s T) T) (T, error) {
	var k8spacketIps = service.k8sClient.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))

	var in T
	out := t

	for _, ip := range k8spacketIps {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(url, ip), nil)
		resp, err := service.httpClient.Do(req)

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
