package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchArticle_UserAgent(t *testing.T) {
	t.Parallel()
	const userAgent = "literature-test/1.0"
	userAgents := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			userAgents <- request.Header.Get("User-Agent")
			_, _ = fmt.Fprint(writer, identityTestArticleXML)
		},
	))
	defer server.Close()
	service := NewArticleService(
		WithArticleHTTPClient(server.Client()),
		WithArticleBaseURL(server.URL),
		WithArticleUserAgent(userAgent),
	)

	_, err := service.FetchArticle("123")
	require.NoError(t, err)
	require.Equal(t, userAgent, <-userAgents)
}

func TestSearchPubMed_UserAgent(t *testing.T) {
	t.Parallel()
	const userAgent = "literature-search-test/1.0"
	userAgents := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			userAgents <- request.Header.Get("User-Agent")
			_, _ = fmt.Fprint(writer, "<eSearchResult></eSearchResult>")
		},
	))
	defer server.Close()
	service := NewSearchService(
		WithSearchHTTPClient(server.Client()),
		WithSearchBaseURL(server.URL),
		WithSearchUserAgent(userAgent),
	)

	_, err := service.SearchPubMed("biology", 5, 2)
	require.NoError(t, err)
	require.Equal(t, userAgent, <-userAgents)
}

func TestPDFService_UsesInjectedArticleService(t *testing.T) {
	t.Parallel()
	const (
		apiValue = "pdf-injected-identity"
		tool     = "pdf-injected-tool"
		email    = "pdf@example.org"
	)
	queries := make(chan url.Values, 1)
	mockArticleHandler := mockEfetchHandler()
	efetchServer := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			queries <- request.URL.Query()
			mockArticleHandler.ServeHTTP(writer, request)
		},
	))
	defer efetchServer.Close()
	s3Server := httptest.NewServer(mockS3Handler(mockOKVersions(1), ""))
	defer s3Server.Close()

	articleHTTPClient := efetchServer.Client()
	articleService := NewArticleService(
		WithArticleHTTPClient(articleHTTPClient),
		WithArticleBaseURL(efetchServer.URL),
		WithArticleIdentity(Identity{
			APIKey: apiValue,
			Tool:   tool,
			Email:  email,
		}),
	)
	pdfHTTPClient := s3Server.Client()
	service := NewPDFService(
		WithPDFHTTPClient(pdfHTTPClient),
		WithPDFArticleService(articleService),
	)
	service.pdfBaseURL = s3Server.URL

	available, err := service.IsPDFAvailable(testPMID)
	req := require.New(t)
	req.NoError(err)
	req.True(available)
	req.Same(articleService, service.articleService)
	req.Same(pdfHTTPClient, service.httpClient)
	req.Same(articleHTTPClient, service.articleService.httpClient)
	query := <-queries
	req.Equal(testPMID, query.Get("id"))
	req.Equal(apiValue, query.Get("api_key"))
	req.Equal(tool, query.Get("tool"))
	req.Equal(email, query.Get("email"))
}
