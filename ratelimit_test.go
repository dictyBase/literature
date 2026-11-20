package literature

import (
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dictybase/literature/internal"
	"github.com/stretchr/testify/require"
)

func ncbiLimit(t *testing.T, client *Client) float64 {
	t.Helper()
	transport, ok := client.ncbiHTTPClient().Transport.(*internal.NCBITransport)
	require.True(t, ok)
	return float64(transport.Limiter.Limit())
}

func TestDefaultRateLimit(t *testing.T) {
	t.Parallel()
	withoutKey, err := New(WithBaseURL("https://ratelimit-a.test"))
	require.NoError(t, err)
	require.InDelta(t, 3, ncbiLimit(t, withoutKey), 0)

	withKey, err := New(WithBaseURL("https://ratelimit-b.test"), WithAPIKey("k"))
	require.NoError(t, err)
	require.InDelta(t, 10, ncbiLimit(t, withKey), 0)

	custom, err := New(
		WithBaseURL("https://ratelimit-c.test"),
		WithAPIKey("k"),
		WithRateLimit(8),
	)
	require.NoError(t, err)
	require.InDelta(t, 8, ncbiLimit(t, custom), 0)
}

// TestSeparateClientsShareBudget sends concurrent requests through two
// clients that use the same key. The shared limiter spaces all of them.
func TestSeparateClientsShareBudget(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))
	defer server.Close()
	newClient := func() *http.Client {
		client, err := New(
			WithHTTPClient(server.Client()),
			WithBaseURL(server.URL),
			WithAPIKey("shared-budget"),
			WithRateLimit(20),
		)
		require.NoError(t, err)
		return client.ncbiHTTPClient()
	}
	clients := []*http.Client{newClient(), newClient()}

	const requests = 6
	statuses := make(chan int, requests)
	start := time.Now()
	var group sync.WaitGroup
	for i := range requests {
		group.Go(func() {
			resp, err := clients[i%2].Get(server.URL)
			if err != nil {
				statuses <- 0
				return
			}
			_ = resp.Body.Close()
			statuses <- resp.StatusCode
		})
	}
	group.Wait()
	elapsed := time.Since(start)
	close(statuses)
	succeeded := 0
	for status := range statuses {
		if status == http.StatusOK {
			succeeded++
		}
	}
	require.Equal(t, requests, succeeded, "expect every request to succeed")
	// 6 requests at 20 per second across both clients: 5 x 50 ms.
	require.GreaterOrEqual(t, elapsed, 240*time.Millisecond)
}

type countingTransport struct {
	calls atomic.Int32
}

func (c *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.calls.Add(1)
	return http.DefaultTransport.RoundTrip(req)
}

func TestCustomTransportStaysInPlace(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("testdata/efetch-single.xml")
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(body)
		},
	))
	defer server.Close()
	custom := &countingTransport{}
	httpClient := &http.Client{Transport: custom}
	client, err := New(WithHTTPClient(httpClient), WithBaseURL(server.URL))
	require.NoError(t, err)

	_, err = client.GetArticle("12345")
	require.NoError(t, err)
	require.Equal(t, int32(1), custom.calls.Load(), "the custom transport sends the request")
	require.Same(t, custom, httpClient.Transport, "the caller's client is not changed")
}

func TestRateAndRetryOptionsRejectNegative(t *testing.T) {
	t.Parallel()
	_, err := New(WithRateLimit(-1))
	require.Error(t, err)
	_, err = New(WithRetryPolicy(-1, time.Second))
	require.Error(t, err)
	_, err = New(WithRateLimit(math.NaN()))
	require.Error(t, err)
	_, err = New(WithRateLimit(math.Inf(1)))
	require.Error(t, err)
}

func TestGetArticleRetriesAfter429(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("testdata/efetch-single.xml")
	require.NoError(t, err)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_, _ = w.Write(body)
		},
	))
	defer server.Close()
	client, err := New(
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithRetryPolicy(3, time.Millisecond),
	)
	require.NoError(t, err)

	article, err := client.GetArticle("12345")
	require.NoError(t, err)
	require.NotNil(t, article)
	require.Equal(t, int32(2), calls.Load())
}

func TestGetArticleReportsExhausted429(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		},
	))
	defer server.Close()
	client, err := New(
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithRetryPolicy(1, time.Millisecond),
	)
	require.NoError(t, err)

	_, err = client.GetArticle("12345")
	require.ErrorContains(t, err, "efetch returned status 429")
}
