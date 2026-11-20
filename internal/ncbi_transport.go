package internal

import (
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// NCBITransport limits the request rate to NCBI E-utilities and retries a
// request that NCBI answers with 429 Too Many Requests.
type NCBITransport struct {
	// Base sends the request. A nil Base uses http.DefaultTransport.
	Base http.RoundTripper
	// Limiter spaces the requests. Each attempt, retries included, waits
	// for a token.
	Limiter *rate.Limiter
	// MaxRetries is the number of retries after a 429. Zero disables retries.
	MaxRetries int
	// RetryDelay is the wait before a retry when the 429 response has no
	// usable Retry-After header.
	RetryDelay time.Duration
}

// RoundTrip implements http.RoundTripper.
func (t *NCBITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		if err := t.Limiter.Wait(req.Context()); err != nil {
			return nil, err
		}
		resp, err := t.base().RoundTrip(req)
		if err != nil ||
			resp.StatusCode != http.StatusTooManyRequests ||
			attempt >= t.MaxRetries ||
			!canResend(req) {
			return resp, err
		}
		wait := retryAfter(resp, t.RetryDelay)
		// Drain a bounded part of the body so the connection can be reused.
		_, _ = io.CopyN(io.Discard, resp.Body, maxDrainBytes)
		_ = resp.Body.Close()
		timer := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}
}

func (t *NCBITransport) base() http.RoundTripper {
	if t.Base == nil {
		return http.DefaultTransport
	}
	return t.Base
}

// canResend reports whether the request can go out again unchanged. The
// E-utilities requests of this package are GET requests without a body.
func canResend(req *http.Request) bool {
	return req.Body == nil || req.Body == http.NoBody
}

// maxDrainBytes bounds the read of a rejected 429 body.
const maxDrainBytes = 64 << 10

// retryAfter reads a Retry-After header, in seconds or as an HTTP date, and
// falls back to the given delay.
func retryAfter(resp *http.Response, fallback time.Duration) time.Duration {
	value := resp.Header.Get("Retry-After")
	if secs, err := strconv.Atoi(value); err == nil {
		if secs <= 0 {
			return fallback
		}
		return time.Duration(secs) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil {
		if wait := time.Until(at); wait > 0 {
			return wait
		}
	}
	return fallback
}

type limiterKey struct {
	baseURL string
	apiKey  string
}

var (
	limitersMu sync.Mutex
	limiters   = map[limiterKey]*rate.Limiter{}
)

// SharedLimiter returns the limiter for one upstream URL and API key in this
// process. NCBI counts requests per key, or per address without a key, so
// all clients in the process with the same key share one budget. When
// clients ask for different rates, the lowest rate applies to all of them.
// The burst is one, so the limiter spaces requests evenly at
// requestsPerSecond. It does not coordinate with other processes that use
// the same key.
func SharedLimiter(
	baseURL, apiKey string,
	requestsPerSecond float64,
) *rate.Limiter {
	key := limiterKey{baseURL: baseURL, apiKey: apiKey}
	limitersMu.Lock()
	defer limitersMu.Unlock()
	if lim, ok := limiters[key]; ok {
		if rate.Limit(requestsPerSecond) < lim.Limit() {
			lim.SetLimit(rate.Limit(requestsPerSecond))
		}
		return lim
	}
	lim := rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
	limiters[key] = lim
	return lim
}
