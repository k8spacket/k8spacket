package tlsparser

import (
	"encoding/json"
	"fmt"
	"github.com/k8spacket/k8s-api/v2"
	"github.com/k8spacket/k8spacket/modules/db"
	"github.com/k8spacket/k8spacket/modules/tls-parser/certificate"
	tls_parser_log "github.com/k8spacket/k8spacket/modules/tls-parser/log"
	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
	"github.com/k8spacket/k8spacket/modules/tls-parser/repository"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Service struct {
	repo repository.IRepository
}

func (service *Service) StoreInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails) {
	var id = strconv.Itoa(int(db.HashId(fmt.Sprintf("%s-%s", tlsConnection.Src, tlsConnection.Dst))))
	tlsConnection.Id = id
	service.repo.UpsertConnection(id, tlsConnection)
	tlsDetails.Id = id
	service.repo.UpsertDetails(id, tlsDetails, certificate.UpdateCertificateInfo)
}

func (service *Service) getConnection(id string) model.TLSDetails {
	return service.repo.Read(id)
}

func (service *Service) filterConnections(query url.Values) []model.TLSConnection {
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
	return service.repo.Query(rangeFrom, rangeTo)
}

func (service *Service) buildConnectionsResponse(url string) ([]model.TLSConnection, error) {
	resultFunc := func(destination, source []model.TLSConnection) []model.TLSConnection {
		return append(destination, source...)
	}
	return buildResponse(url, []model.TLSConnection{}, resultFunc)
}

func (service *Service) buildDetailsResponse(url string) (model.TLSDetails, error) {
	resultFunc := func(destination, source model.TLSDetails) model.TLSDetails {
		if !reflect.DeepEqual(source, model.TLSDetails{}) {
			return source
		} else {
			return destination
		}
	}
	return buildResponse(url, model.TLSDetails{}, resultFunc)
}

func buildResponse[T model.TLSDetails | []model.TLSConnection](url string, t T, resultFunc func(d T, s T) T) (T, error) {
	var k8spacketIps = k8s.GetPodIPsBySelectors(os.Getenv("K8S_PACKET_API_FIELD_SELECTOR"), os.Getenv("K8S_PACKET_API_LABEL_SELECTOR"))

	var in T
	out := t

	for _, ip := range k8spacketIps {
		resp, err := http.Get(fmt.Sprintf(url, ip))

		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot get stats: %+v", err)
			return out, err
		}

		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot read stats response: %+v", err)
			return out, err
		}

		_ = json.Unmarshal(responseData, &in)
		if err != nil {
			tls_parser_log.LOGGER.Printf("[api] Cannot parse stats response: %+v", err)
			return out, err
		}

		out = resultFunc(out, in)
	}

	return out, nil
}
