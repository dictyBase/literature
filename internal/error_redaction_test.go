package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const redactionIntegrationValue = "identity-redaction-fixture"

func newClosedArticleService() *ArticleService {
	server := httptest.NewServer(http.NotFoundHandler())
	service := NewArticleService(
		WithArticleHTTPClient(server.Client()),
		WithArticleBaseURL(server.URL),
		WithArticleIdentity(Identity{APIKey: redactionIntegrationValue}),
	)
	server.Close()
	return service
}

func newClosedSearchService() *SearchService {
	server := httptest.NewServer(http.NotFoundHandler())
	service := NewSearchService(
		WithSearchHTTPClient(server.Client()),
		WithSearchBaseURL(server.URL),
		WithSearchIdentity(Identity{APIKey: redactionIntegrationValue}),
	)
	server.Close()
	return service
}

func TestFetchArticle_ErrorDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()
	_, err := newClosedArticleService().FetchArticle("123")

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), redactionIntegrationValue)
}

func TestSearchPubMed_ErrorDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()
	_, err := newClosedSearchService().SearchPubMed("biology", 10, 0)

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), redactionIntegrationValue)
}

func TestFetchPubMedDetails_ErrorDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()
	_, err := newClosedSearchService().FetchPubMedDetails("env", "2", 10, 0)

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), redactionIntegrationValue)
}

func TestFetchArticle_InvalidRequestURLDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()
	service := NewArticleService(
		WithArticleBaseURL("http://%zz"),
		WithArticleIdentity(Identity{APIKey: redactionIntegrationValue}),
	)
	_, err := service.FetchArticle("123")

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), redactionIntegrationValue)
}

func TestSearchPubMed_InvalidRequestURLDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()
	service := NewSearchService(
		WithSearchBaseURL("http://%zz"),
		WithSearchIdentity(Identity{APIKey: redactionIntegrationValue}),
	)
	_, err := service.SearchPubMed("biology", 10, 0)

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), redactionIntegrationValue)
}

func TestFetchPubMedDetails_InvalidRequestURLDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()
	service := NewSearchService(
		WithSearchBaseURL("http://%zz"),
		WithSearchIdentity(Identity{APIKey: redactionIntegrationValue}),
	)
	_, err := service.FetchPubMedDetails("env", "2", 10, 0)

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), redactionIntegrationValue)
}
