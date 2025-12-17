package httpclient

import "net/http"

type HttpClient struct {
	Client
}

func (httpClient *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
