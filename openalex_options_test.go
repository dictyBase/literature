package literature

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithOpenAlexEmail_PolitePool(t *testing.T) {
	t.Parallel()

	client, err := NewOpenAlexClient(WithOpenAlexEmail("polite@example.com"))

	req := require.New(t)
	req.NoError(err)
	req.Equal("polite@example.com", client.email)
}

func TestWithOpenAlexAPIKey(t *testing.T) {
	t.Parallel()

	client, err := NewOpenAlexClient(WithOpenAlexAPIKey("secret-key"))

	req := require.New(t)
	req.NoError(err)
	req.Equal("secret-key", client.apiKey)
}

func TestWithOpenAlexHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 5 * time.Second}
	client, err := NewOpenAlexClient(WithOpenAlexHTTPClient(custom))

	req := require.New(t)
	req.NoError(err)
	req.Equal(custom, client.httpClient)
}

func TestWithOpenAlexTimeout(t *testing.T) {
	t.Parallel()

	client, err := NewOpenAlexClient(WithOpenAlexTimeout(7 * time.Second))

	req := require.New(t)
	req.NoError(err)
	req.Equal(7*time.Second, client.httpClient.Timeout)
}

func TestWithOpenAlexBaseURL(t *testing.T) {
	t.Parallel()

	client, err := NewOpenAlexClient(WithOpenAlexBaseURL("http://localhost:9999"))

	req := require.New(t)
	req.NoError(err)
	req.Equal("http://localhost:9999", client.baseURL)
}

func TestWithOpenAlexUserAgent(t *testing.T) {
	t.Parallel()

	client, err := NewOpenAlexClient(WithOpenAlexUserAgent("custom-agent"))

	req := require.New(t)
	req.NoError(err)
	req.Equal("custom-agent", client.userAgent)
}

func TestNewOpenAlexClient_Defaults(t *testing.T) {
	t.Parallel()

	client, err := NewOpenAlexClient()

	req := require.New(t)
	req.NoError(err)
	req.Equal("https://api.openalex.org", client.baseURL)
	req.Equal("openalex-go-client/1.0", client.userAgent)
	req.Equal(30*time.Second, client.httpClient.Timeout)
	req.Empty(client.email)
	req.Empty(client.apiKey)
}

func TestOpenAlexSearchOptions(t *testing.T) {
	t.Parallel()

	config := &openAlexSearchConfig{}
	WithOpenAlexPerPage(50)(config)
	WithOpenAlexCursor("cursor-1")(config)
	WithOpenAlexSort("publication_date:desc")(config)

	req := require.New(t)
	req.Equal(50, config.perPage)
	req.Equal("cursor-1", config.cursor)
	req.Equal("publication_date:desc", config.sort)
}

func TestOpenAlexSearchOptions_IgnoreZeroValues(t *testing.T) {
	t.Parallel()

	config := &openAlexSearchConfig{}
	WithOpenAlexPerPage(-1)(config)
	WithOpenAlexCursor("")(config)
	WithOpenAlexSort("")(config)

	req := require.New(t)
	req.Equal(0, config.perPage)
	req.Empty(config.cursor)
	req.Empty(config.sort)
}
