package httpclient

import (
	"net/http"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
