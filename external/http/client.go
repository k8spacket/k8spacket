package httpclient

import "net/http"

type HttpClient struct {
	IHttpClient
}

func (httpClient *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
