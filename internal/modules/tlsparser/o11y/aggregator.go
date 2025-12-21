package o11y

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	httpclient "github.com/k8spacket/k8spacket/internal/thirdparty/http"
)

// aggregateTLSResponses fetches TLS responses from peer k8spacket pods concurrently and merges them.
func aggregateTLSResponses[T model.TLSDetails | []model.TLSConnection](ctx context.Context, podIPs []string, urlTemplate string, client httpclient.Client, zero T, merge func(dst T, src T) T) (T, []error) {
	if len(podIPs) == 0 {
		return zero, nil
	}

	const maxConcurrent = 5
	const requestTimeout = 5 * time.Second

	type result struct {
		value T
		err   error
	}

	sem := make(chan struct{}, maxConcurrent)
	resCh := make(chan result, len(podIPs))
	var wg sync.WaitGroup

	for _, ip := range podIPs {
		ip := ip
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf(urlTemplate, ip), nil)
			if err != nil {
				slog.Error("[api] Cannot get stats", "Error", err)
				resCh <- result{err: err}
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				slog.Error("[api] Cannot get stats", "Error", err)
				resCh <- result{err: err}
				return
			}
			if resp.Body == nil {
				err = fmt.Errorf("peer %s returned empty body", ip)
				slog.Error("[api] Cannot read stats response", "Error", err)
				resCh <- result{err: err}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("peer %s status %d", ip, resp.StatusCode)
				slog.Error("[api] Cannot get stats", "Error", err)
				resCh <- result{err: err}
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				slog.Error("[api] Cannot read stats response", "Error", err)
				resCh <- result{err: err}
				return
			}

			var fetched T
			if err := json.Unmarshal(body, &fetched); err != nil {
				slog.Error("[api] Cannot parse stats response", "Error", err)
				resCh <- result{err: err}
				return
			}

			resCh <- result{value: fetched}
		}()
	}

	wg.Wait()
	close(resCh)

	out := zero
	var errs []error
	for res := range resCh {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		out = merge(out, res.value)
	}

	if len(errs) > 0 {
		slog.Warn("[api] partial tlsparser aggregation failures", "errors", errs)
	}

	return out, errs
}
