package literature

import (
	"math"
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

// WithAPIKey sets the NCBI API key sent on every E-utilities request.
// Empty by default: callers should provide their own key.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) error {
		c.apiKey = apiKey
		return nil
	}
}

// WithTool sets the tool parameter identifying the software making requests.
func WithTool(tool string) Option {
	return func(client *Client) error {
		if err := client.validate.Var(tool, "omitempty,excludesall= "); err != nil {
			return &Error{
				Type:    ErrorTypeInvalidInput,
				Message: "tool must not contain spaces",
			}
		}
		client.tool = tool
		return nil
	}
}

// WithEmail sets the developer contact email sent on every E-utilities request.
func WithEmail(email string) Option {
	return func(client *Client) error {
		if err := client.validate.Var(email, "omitempty,email"); err != nil {
			return &Error{
				Type:    ErrorTypeInvalidInput,
				Message: "email must be a valid address",
			}
		}
		client.email = email
		return nil
	}
}

// WithRetryPolicy sets the retry of an E-utilities request that NCBI answers
// with 429 Too Many Requests. maxRetries is the number of retries, and 0
// turns retries off. retryDelay is the wait before a retry when the response
// has no Retry-After header. The defaults are 3 retries and 1 second.
func WithRetryPolicy(maxRetries int, retryDelay time.Duration) Option {
	return func(client *Client) error {
		if maxRetries < 0 || retryDelay < 0 {
			return &Error{
				Type:    ErrorTypeInvalidInput,
				Message: "retry policy values must not be negative",
			}
		}
		client.maxRetries = maxRetries
		if retryDelay > 0 {
			client.retryDelay = retryDelay
		}
		return nil
	}
}

// WithRateLimit sets the number of E-utilities requests per second. All
// clients in the process with the same base URL and API key share one
// limiter, and the lowest requested rate applies to all of them. The default
// is 10 with an API key and 3 without one, the NCBI limits. The limiter does
// not see other processes, so use a lower value when they use the same key.
func WithRateLimit(requestsPerSecond float64) Option {
	return func(client *Client) error {
		if requestsPerSecond < 0 ||
			math.IsNaN(requestsPerSecond) ||
			math.IsInf(requestsPerSecond, 0) {
			return &Error{
				Type:    ErrorTypeInvalidInput,
				Message: "rate limit must be a finite value that is not negative",
			}
		}
		client.rateLimit = requestsPerSecond
		return nil
	}
}
