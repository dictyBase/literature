package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchPubMed_IdentityParams(t *testing.T) {
	t.Parallel()
	const (
		apiKey = "search-test-key"
		tool   = "literature-client"
		email  = "dev@example.org"
		query  = "cancer & gene expression"
	)
	requests := make(chan url.Values, 2)
	mux := http.NewServeMux()
	mux.HandleFunc("/esearch.fcgi", func(writer http.ResponseWriter, request *http.Request) {
		requests <- request.URL.Query()
		_, _ = fmt.Fprint(writer, "<eSearchResult></eSearchResult>")
	})
	mux.HandleFunc("/efetch.fcgi", func(writer http.ResponseWriter, request *http.Request) {
		requests <- request.URL.Query()
		_, _ = fmt.Fprint(writer, "<PubmedArticleSet></PubmedArticleSet>")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	service := &SearchService{
		httpClient: server.Client(),
		baseURL:    server.URL,
		identity: Identity{
			APIKey: apiKey,
			Tool:   tool,
			Email:  email,
		},
	}
	req := require.New(t)
	_, err := service.SearchPubMed(query, 7, 3)
	req.NoError(err)
	_, err = service.FetchPubMedDetails("web env", "2", 7, 3)
	req.NoError(err)

	searchQuery := <-requests
	req.Equal(query, searchQuery.Get("term"))
	req.Equal("7", searchQuery.Get("retmax"))
	req.Equal("3", searchQuery.Get("retstart"))
	req.Equal("y", searchQuery.Get("usehistory"))
	req.Equal(apiKey, searchQuery.Get("api_key"))
	req.Equal(tool, searchQuery.Get("tool"))
	req.Equal(email, searchQuery.Get("email"))

	efetchQuery := <-requests
	req.Equal("web env", efetchQuery.Get("WebEnv"))
	req.Equal("2", efetchQuery.Get("query_key"))
	req.Equal("7", efetchQuery.Get("retmax"))
	req.Equal("3", efetchQuery.Get("retstart"))
	req.Equal(apiKey, efetchQuery.Get("api_key"))
	req.Equal(tool, efetchQuery.Get("tool"))
	req.Equal(email, efetchQuery.Get("email"))
}
