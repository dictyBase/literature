package literature

import (
	"net/http"
	"time"
)

// OpenAlexOption configures the OpenAlex client.
type OpenAlexOption func(*OpenAlexClient) error

// WithOpenAlexHTTPClient sets a custom HTTP client for the OpenAlex client.
func WithOpenAlexHTTPClient(client *http.Client) OpenAlexOption {
	return func(c *OpenAlexClient) error {
		if client != nil {
			c.httpClient = client
		}
		return nil
	}
}

// WithOpenAlexTimeout sets the HTTP client timeout.
func WithOpenAlexTimeout(timeout time.Duration) OpenAlexOption {
	return func(c *OpenAlexClient) error {
		if timeout > 0 {
			if c.httpClient == nil {
				c.httpClient = &http.Client{}
			}
			c.httpClient.Timeout = timeout
		}
		return nil
	}
}

// WithOpenAlexBaseURL sets a custom base URL for the OpenAlex API.
// This is primarily useful for testing with mock servers.
func WithOpenAlexBaseURL(baseURL string) OpenAlexOption {
	return func(c *OpenAlexClient) error {
		if baseURL != "" {
			c.baseURL = baseURL
		}
		return nil
	}
}

// WithOpenAlexUserAgent sets a custom User-Agent header for HTTP requests.
func WithOpenAlexUserAgent(userAgent string) OpenAlexOption {
	return func(c *OpenAlexClient) error {
		if userAgent != "" {
			c.userAgent = userAgent
		}
		return nil
	}
}

// WithOpenAlexEmail sets the email contact for the OpenAlex "Polite Pool".
// The email is sent as the mailto query parameter on every request and
// entitles the client to faster, more reliable rate limits for free.
func WithOpenAlexEmail(email string) OpenAlexOption {
	return func(c *OpenAlexClient) error {
		if email != "" {
			c.email = email
		}
		return nil
	}
}

// WithOpenAlexAPIKey sets the OpenAlex Premium API key. The key is sent as
// the api_key query parameter and is automatically redacted from error
// strings produced by failed HTTP requests.
func WithOpenAlexAPIKey(apiKey string) OpenAlexOption {
	return func(c *OpenAlexClient) error {
		if apiKey != "" {
			c.apiKey = apiKey
		}
		return nil
	}
}

// OpenAlexSearchOption configures a works list request, such as citation
// graph queries.
type OpenAlexSearchOption func(*openAlexSearchConfig)

// openAlexSearchConfig holds pagination and sorting options for works
// list requests.
type openAlexSearchConfig struct {
	perPage int
	cursor  string
	sort    string
}

// WithOpenAlexPerPage sets the number of works per page (maximum 200).
func WithOpenAlexPerPage(perPage int) OpenAlexSearchOption {
	return func(config *openAlexSearchConfig) {
		if perPage > 0 {
			config.perPage = perPage
		}
	}
}

// WithOpenAlexCursor sets the pagination cursor for the next page.
func WithOpenAlexCursor(cursor string) OpenAlexSearchOption {
	return func(config *openAlexSearchConfig) {
		if cursor != "" {
			config.cursor = cursor
		}
	}
}

// WithOpenAlexSort sets the sort expression, for example
// "cited_by_count:desc".
func WithOpenAlexSort(sort string) OpenAlexSearchOption {
	return func(config *openAlexSearchConfig) {
		if sort != "" {
			config.sort = sort
		}
	}
}
