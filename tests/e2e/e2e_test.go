package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var client = os.Getenv("CLIENT_IP")
var host = os.Getenv("HOST_IP")
var guest = os.Getenv("GUEST_IP")
var port = 16676

func TestNodegraphHeathEndpoint(t *testing.T) {

	assert.Eventually(t, func() bool {
		httpClient := &http.Client{}
		httpClient.Timeout = 3 * time.Second
		resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/nodegraph/api/health", host, port))
		if err != nil {
			log.Println(err)
			return false
		}
		return assert.EqualValues(t, resp.StatusCode, http.StatusOK)
	}, 10*time.Second, 1*time.Second)

}

func TestNodegraphFieldsEndpoint(t *testing.T) {

	var tests = []struct {
		scenario string
		wantFile string
	}{
		{"", "./resources/fields_connection.json"},
		{"connection", "./resources/fields_connection.json"},
		{"bytes", "./resources/fields_bytes.json"},
		{"duration", "./resources/fields_duration.json"},
	}

	httpClient := &http.Client{}
	httpClient.Timeout = 1 * time.Second

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()

			want, _ := os.ReadFile(test.wantFile)

			assert.Eventually(t, func() bool {
				httpClient.Timeout = 3 * time.Second

				req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/nodegraph/api/graph/fields?stats-type=%s", host, port, test.scenario), nil)
				req.Header.Set("Connection", "close")
				resp, err := httpClient.Do(req)
				if err != nil {
					log.Println(err)
					return false
				}
				body, _ := io.ReadAll(resp.Body)

				return assert.EqualValues(t, resp.StatusCode, http.StatusOK) && assert.EqualValues(t, strings.TrimSpace(string(want)), strings.TrimSpace(string(body)))
			}, 10*time.Second, 1*time.Second)

		})
	}
}

func TestNodegraphDataConnectionEndpoint(t *testing.T) {

	initData()

	doNodegraphTest(t, "connection", func(nodeMainStatVal string, nodeSecStatVal string, nodeArg1Val float64, nodeArg2Val float64, nodeArg3Val float64, edgeMainStatVal string, edgeSecStatVal string) bool {
		re := regexp.MustCompile("\\w: (\\d*).*")

		nodeAll, _ := strconv.ParseInt(re.FindStringSubmatch(nodeMainStatVal)[1], 10, 64)

		edgeAll, _ := strconv.ParseInt(re.FindStringSubmatch(edgeMainStatVal)[1], 10, 64)

		return assert.Greater(t, nodeAll, int64(0)) &&
			assert.Greater(t, edgeAll, int64(0)) &&
			assert.Greater(t, nodeArg2Val, 0.0)
	})
}

func TestNodegraphDataBytesEndpoint(t *testing.T) {

	doNodegraphTest(t, "bytes", func(nodeMainStatVal string, nodeSecStatVal string, nodeArg1Val float64, nodeArg2Val float64, nodeArg3Val float64, edgeMainStatVal string, edgeSecStatVal string) bool {
		re := regexp.MustCompile("\\w: (\\d*\\.\\d*).*")

		nodeRecv, _ := strconv.ParseFloat(re.FindStringSubmatch(nodeMainStatVal)[1], 64)
		nodeResp, _ := strconv.ParseFloat(re.FindStringSubmatch(nodeSecStatVal)[1], 64)

		edgeSent, _ := strconv.ParseFloat(re.FindStringSubmatch(edgeMainStatVal)[1], 64)
		edgeRecv, _ := strconv.ParseFloat(re.FindStringSubmatch(edgeSecStatVal)[1], 64)

		return assert.Greater(t, nodeRecv, 0.0) &&
			assert.Greater(t, nodeResp, 0.0) &&
			assert.Greater(t, edgeSent, 0.0) &&
			assert.Greater(t, edgeRecv, 0.0) &&
			assert.Greater(t, nodeArg1Val, 0.0) &&
			assert.Greater(t, nodeArg2Val, 0.0)
	})
}

func TestNodegraphDataDurationEndpoint(t *testing.T) {

	doNodegraphTest(t, "duration", func(nodeMainStatVal string, nodeSecStatVal string, nodeArg1Val float64, nodeArg2Val float64, nodeArg3Val float64, edgeMainStatVal string, edgeSecStatVal string) bool {
		nodeAvg, _ := time.ParseDuration(nodeMainStatVal[5:])
		nodeMax, _ := time.ParseDuration(nodeSecStatVal[5:])
		edgeAvg, _ := time.ParseDuration(edgeMainStatVal[5:])
		edgeMax, _ := time.ParseDuration(edgeSecStatVal[5:])

		return assert.Greater(t, nodeAvg, time.Second*0) &&
			assert.Greater(t, nodeMax, time.Second*0) &&
			assert.Greater(t, edgeAvg, time.Second*0) &&
			assert.Greater(t, edgeMax, time.Second*0) &&
			assert.Greater(t, nodeArg1Val, 0.0) &&
			assert.Greater(t, nodeArg2Val, 0.0)
	})

}

func TestTlsParserTLS13Endpoint(t *testing.T) {
	doTlsParserTest(t, "k8spacket-tls13.domain", "TLS 1.3")
}

func TestTlsParserTLS12Endpoint(t *testing.T) {
	doTlsParserTest(t, "k8spacket-tls12.domain", "TLS 1.2")
}

func doTlsParserTest(t *testing.T, domain string, tlsVer string) {
	httpClient := &http.Client{}
	httpClient.Timeout = 3 * time.Second
	detailId := ""
	tlsVersion := ""
	cipher := ""
	assert.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/tlsparser/api/data", host, port), nil)
		req.Header.Set("Connection", "close")
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println(err)
			return false
		}
		body, _ := io.ReadAll(resp.Body)

		detailId = gjson.GetBytes(body, fmt.Sprintf("#(domain==\"%s\").id", domain)).String()
		tlsVersion = gjson.GetBytes(body, fmt.Sprintf("#(domain==\"%s\").usedTLSVersion", domain)).String()
		cipher = gjson.GetBytes(body, fmt.Sprintf("#(domain==\"%s\").usedCipherSuite", domain)).String()
		lastSeenStr := gjson.GetBytes(body, fmt.Sprintf("#(domain==\"%s\").lastSeen", domain)).String()
		lastSeen, _ := time.Parse("2006-01-02T15:04:05.000000000Z", lastSeenStr)

		return tlsVer == tlsVersion &&
			cipher != "" &&
			time.Now().After(lastSeen)
	}, 10*time.Second, 1*time.Second)

	assert.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/tlsparser/api/data/%s", host, port, detailId), nil)
		req.Header.Set("Connection", "close")
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println(err)
			return false
		}
		body, _ := io.ReadAll(resp.Body)

		d := gjson.GetBytes(body, "domain").String()
		clientTLSVersions := gjson.GetBytes(body, "clientTLSVersions").String()
		clientCipherSuites := gjson.GetBytes(body, "clientCipherSuites").String()

		notBeforeStr := gjson.GetBytes(body, "certificate.notBefore").String()
		notAfterStr := gjson.GetBytes(body, "certificate.notAfter").String()

		notBefore, _ := time.Parse("2006-01-02T15:04:05Z", notBeforeStr)
		notAfter, _ := time.Parse("2006-01-02T15:04:05Z", notAfterStr)

		chain := gjson.GetBytes(body, "certificate.serverChain").String()

		return domain == d &&
			strings.Contains(clientTLSVersions, tlsVersion) &&
			strings.Contains(clientCipherSuites, cipher) &&
			time.Now().After(notBefore) &&
			time.Now().Before(notAfter) &&
			strings.Contains(chain, "C=PL,ST=Poznan,UnknownOID=2.5.4.7,O=k8spacket,OU=k8spacket,CN=k8spacket.domain")
	}, 10*time.Second, 1*time.Second)
}

func doNodegraphTest(t *testing.T, statsType string, assertFunc func(nodeMainStatVal string, nodeSecStatVal string, nodeArg1Val float64, nodeArg2Val float64, nodeArg3Val float64, edgeMainStatVal string, edgeSecStatVal string) bool) {
	httpClient := &http.Client{}
	httpClient.Timeout = 3 * time.Second
	assert.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/nodegraph/api/graph/data?stats-type=%s", host, port, statsType), nil)
		req.Header.Set("Connection", "close")
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println(err)
			return false
		}
		body, _ := io.ReadAll(resp.Body)

		nodeMainStatVal := gjson.GetBytes(body, fmt.Sprintf("nodes.#(id==\"%s\").mainStat", guest)).String()
		nodeSecStatVal := gjson.GetBytes(body, fmt.Sprintf("nodes.#(id==\"%s\").secondaryStat", guest)).String()
		nodeArg1Val := gjson.GetBytes(body, fmt.Sprintf("nodes.#(id==\"%s\").arc__1", guest)).Float()
		nodeArg2Val := gjson.GetBytes(body, fmt.Sprintf("nodes.#(id==\"%s\").arc__2", guest)).Float()
		nodeArg3Val := gjson.GetBytes(body, fmt.Sprintf("nodes.#(id==\"%s\").arc__3", guest)).Float()

		edgeMainStatVal := gjson.GetBytes(body, fmt.Sprintf("edges.#(id==\"%s-%s\").mainStat", client, guest)).String()
		edgeSecStatVal := gjson.GetBytes(body, fmt.Sprintf("edges.#(id==\"%s-%s\").secondaryStat", client, guest)).String()

		return assertFunc(nodeMainStatVal, nodeSecStatVal, nodeArg1Val, nodeArg2Val, nodeArg3Val, edgeMainStatVal, edgeSecStatVal)

	}, 10*time.Second, 1*time.Second)
}

func initData() {

	var data = []struct {
		size  int
		sleep int
	}{
		{1, 0},
		{1, 2},
		{500, 0},
		{1000, 0},
	}

	body := make([]byte, 1000)
	rand.Read(body)

	httpClient := &http.Client{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient.Timeout = 3 * time.Second

	for _, item := range data {
		req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s:10443?size=%d&sleep=%d", host, item.size, item.sleep), bytes.NewReader(body))
		req.Header.Set("Connection", "close")
		httpClient.Do(req)
	}
}
