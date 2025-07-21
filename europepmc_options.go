package literature

import (
	"net/http"
	"time"
)

// EuropePMCOption configures the EuropePMC client.
type EuropePMCOption func(*EuropePMCClient) error

// WithEuropePMCHTTPClient sets a custom HTTP client for the EuropePMC client.
func WithEuropePMCHTTPClient(client *http.Client) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if client != nil {
			c.httpClient = client
		}
		return nil
	}
}

// WithEuropePMCTimeout sets the HTTP client timeout.
func WithEuropePMCTimeout(timeout time.Duration) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if timeout > 0 {
			if c.httpClient == nil {
				c.httpClient = &http.Client{}
			}
			c.httpClient.Timeout = timeout
		}
		return nil
	}
}

// WithEuropePMCBaseURL sets a custom base URL for the EuropePMC API.
// This is primarily useful for testing with mock servers.
func WithEuropePMCBaseURL(url string) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if url != "" {
			c.baseURL = url
		}
		return nil
	}
}

// WithEuropePMCUserAgent sets a custom User-Agent header for HTTP requests.
func WithEuropePMCUserAgent(userAgent string) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if userAgent != "" {
			c.userAgent = userAgent
		}
		return nil
	}
}

// WithEuropePMCRetryPolicy sets retry behavior for failed requests.
// maxRetries specifies the maximum number of retry attempts.
// retryDelay specifies the delay between retries.
func WithEuropePMCRetryPolicy(maxRetries int, retryDelay time.Duration) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if maxRetries >= 0 {
			c.maxRetries = maxRetries
		}
		if retryDelay > 0 {
			c.retryDelay = retryDelay
		}
		return nil
	}
}

// WithEuropePMCRateLimit sets rate limiting for API requests.
// requestsPerSecond specifies the maximum number of requests per second.
func WithEuropePMCRateLimit(requestsPerSecond float64) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if requestsPerSecond > 0 {
			c.requestsPerSecond = requestsPerSecond
		}
		return nil
	}
}

// WithEuropePMCEmail sets the email contact for API requests.
// EuropePMC recommends providing contact information for high-volume usage.
func WithEuropePMCEmail(email string) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if email != "" {
			c.email = email
		}
		return nil
	}
}

// WithEuropePMCDefaultResultType sets the default result type for searches.
func WithEuropePMCDefaultResultType(resultType string) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if resultType != "" {
			c.defaultResultType = resultType
		}
		return nil
	}
}

// WithEuropePMCDefaultFormat sets the default format for API responses.
func WithEuropePMCDefaultFormat(format string) EuropePMCOption {
	return func(c *EuropePMCClient) error {
		if format != "" {
			c.defaultFormat = format
		}
		return nil
	}
}
