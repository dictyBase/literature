package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// statusSequence answers each request with the next status, and with 200
// after the list ends.
func statusSequence(calls *atomic.Int32, statuses ...int) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		n := int(calls.Add(1)) - 1
		if n < len(statuses) {
			writer.WriteHeader(statuses[n])
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func newTransportClient(maxRetries int) *http.Client {
	return &http.Client{Transport: &NCBITransport{
		Limiter:    rate.NewLimiter(rate.Inf, 1),
		MaxRetries: maxRetries,
		RetryDelay: time.Millisecond,
	}}
}

func TestNCBITransportRetriesAfter429(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls, 429, 429))
	defer server.Close()

	resp, err := newTransportClient(3).Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, int32(3), calls.Load())
}

func TestNCBITransportStopsAfterMaxRetries(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls, 429, 429, 429, 429, 429))
	defer server.Close()

	resp, err := newTransportClient(2).Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	require.Equal(t, int32(3), calls.Load(), "one try and two retries")
}

func TestNCBITransportDoesNotRetryOtherErrors(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls, 500))
	defer server.Close()

	resp, err := newTransportClient(3).Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, int32(1), calls.Load())
}

func TestNCBITransportStopsWaitOnContextCancel(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls, 429))
	defer server.Close()
	client := &http.Client{Transport: &NCBITransport{
		Limiter:    rate.NewLimiter(rate.Inf, 1),
		MaxRetries: 3,
		RetryDelay: time.Hour,
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, int32(1), calls.Load())
}

func TestNCBITransportSpacesRequests(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls))
	defer server.Close()
	client := &http.Client{Transport: &NCBITransport{
		Limiter: rate.NewLimiter(rate.Limit(20), 1),
	}}

	start := time.Now()
	for range 5 {
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		_ = resp.Body.Close()
	}
	// 5 requests at 20 per second: the first is free, four wait 50 ms each.
	require.GreaterOrEqual(t, time.Since(start), 190*time.Millisecond)
}

func TestNCBITransportRetriesUseTokens(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls, 429, 429))
	defer server.Close()
	client := &http.Client{Transport: &NCBITransport{
		Limiter:    rate.NewLimiter(rate.Limit(20), 1),
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	}}

	start := time.Now()
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, int32(3), calls.Load())
	// The two retries each wait for a token: 2 x 50 ms.
	require.GreaterOrEqual(t, time.Since(start), 95*time.Millisecond)
}

func TestNCBITransportCancelsLimiterWait(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(statusSequence(&calls))
	defer server.Close()
	limiter := rate.NewLimiter(rate.Limit(0.1), 1)
	require.True(t, limiter.Allow(), "use the one free token")
	client := &http.Client{Transport: &NCBITransport{Limiter: limiter}}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
	require.Equal(t, int32(0), calls.Load(), "no request while the limiter waits")
}

func TestRetryAfter(t *testing.T) {
	t.Parallel()
	resp := &http.Response{Header: http.Header{}}
	require.Equal(t, time.Second, retryAfter(resp, time.Second))
	resp.Header.Set("Retry-After", "2")
	require.Equal(t, 2*time.Second, retryAfter(resp, time.Second))
	resp.Header.Set("Retry-After", "0")
	require.Equal(t, time.Second, retryAfter(resp, time.Second))
	resp.Header.Set("Retry-After", "Wed, 21 Oct 2015 07:28:00 GMT")
	require.Equal(t, time.Second, retryAfter(resp, time.Second), "a past date")
	resp.Header.Set(
		"Retry-After",
		time.Now().Add(5*time.Second).UTC().Format(http.TimeFormat),
	)
	wait := retryAfter(resp, time.Second)
	require.Greater(t, wait, 3*time.Second)
	require.LessOrEqual(t, wait, 5*time.Second)
}

func TestSharedLimiterPerKey(t *testing.T) {
	t.Parallel()
	// A unique upstream per run, because the registry outlives the test.
	upstream := fmt.Sprintf("https://x-%d.test", time.Now().UnixNano())
	shared := SharedLimiter(upstream, "key-shared", 10)
	require.Same(t, shared, SharedLimiter(upstream, "key-shared", 10))
	require.NotSame(t, shared, SharedLimiter(upstream, "key-b", 10))
	require.NotSame(t, shared, SharedLimiter("https://y.test", "key-shared", 10))
	require.Equal(t, 1, shared.Burst())
	require.InDelta(t, 10, float64(shared.Limit()), 0)

	require.Same(t, shared, SharedLimiter(upstream, "key-shared", 4),
		"shared different rate keeps the shared budget")
	require.InDelta(t, 4, float64(shared.Limit()), 0, "the lowest rate applies")
	SharedLimiter(upstream, "key-shared", 10)
	require.InDelta(t, 4, float64(shared.Limit()), 0, "shared higher rate does not raise it")
}
