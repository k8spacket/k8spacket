package httpclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHttpClient_Do_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	assert.NoError(t, err)

	client := &HttpClient{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}

func TestHttpClient_Do_Error(t *testing.T) {
	// replace DefaultClient with one that always errors
	old := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("transport error")
	})}
	defer func() { http.DefaultClient = old }()

	req, err := http.NewRequest("GET", "http://example.invalid/", nil)
	assert.NoError(t, err)

	client := &HttpClient{}
	resp, err := client.Do(req)
	assert.Error(t, err)
	assert.Contains(t, fmt.Sprintf("%v", err), "transport error")
	assert.Nil(t, resp)
}
