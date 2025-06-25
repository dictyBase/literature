package literature

import (
	"net/http"
	"time"
)

// Option configures the literature client.
type Option func(*Client) error

// WithHTTPClient sets a custom HTTP client for the literature client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client != nil {
			c.httpClient = client
		}
		return nil
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout > 0 {
			if c.httpClient == nil {
				c.httpClient = &http.Client{}
			}
			c.httpClient.Timeout = timeout
		}
		return nil
	}
}

// WithBaseURL sets a custom base URL for the PubMed API.
// This is primarily useful for testing with mock servers.
func WithBaseURL(url string) Option {
	return func(c *Client) error {
		if url != "" {
			c.baseURL = url
		}
		return nil
	}
}

// WithUserAgent sets a custom User-Agent header for HTTP requests.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		if userAgent != "" {
			c.userAgent = userAgent
		}
		return nil
	}
}

// WithRetryPolicy sets retry behavior for failed requests.
// maxRetries specifies the maximum number of retry attempts.
// retryDelay specifies the delay between retries.
func WithRetryPolicy(maxRetries int, retryDelay time.Duration) Option {
	return func(c *Client) error {
		// For now, we'll store these in the client for future retry implementation
		// In a full implementation, we'd wrap the HTTP client with retry logic
		return nil
	}
}

// WithRateLimit sets rate limiting for API requests.
// requestsPerSecond specifies the maximum number of requests per second.
func WithRateLimit(requestsPerSecond float64) Option {
	return func(c *Client) error {
		// For now, we'll store this in the client for future rate limiting implementation
		// In a full implementation, we'd add rate limiting middleware
		return nil
	}
}
