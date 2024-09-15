package main

import (
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
	"time"
)

var host = "127.0.0.1"

func TestNodegraphHeathEndpoint(t *testing.T) {

	assert.Eventually(t, func() bool {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err := http.Get(fmt.Sprintf("http://%s:16676/nodegraph/api/health", host))
		if err != nil {
			log.Fatal(err)
			return false
		}
		return resp.StatusCode == http.StatusOK
	}, time.Second*10, time.Millisecond*1000)

}
