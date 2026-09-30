package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const identityTestArticleXML = `
<PubmedArticleSet>
	<PubmedArticle>
		<MedlineCitation><PMID>123</PMID></MedlineCitation>
	</PubmedArticle>
</PubmedArticleSet>`

func fetchArticleQuery(t *testing.T, identity Identity) url.Values {
	t.Helper()
	queries := make(chan url.Values, 1)
	service, server := newTestArticleService(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			queries <- request.URL.Query()
			_, _ = fmt.Fprint(writer, identityTestArticleXML)
		},
	))
	defer server.Close()
	service.identity = identity

	_, err := service.FetchArticle("123")
	require.NoError(t, err)

	return <-queries
}

func TestFetchArticle_IdentityParams(t *testing.T) {
	t.Parallel()
	const (
		apiKey = "test-api-key"
		tool   = "literature-client"
		email  = "dev@example.org"
	)
	query := fetchArticleQuery(t, Identity{
		APIKey: apiKey,
		Tool:   tool,
		Email:  email,
	})

	require.Equal(t, apiKey, query.Get("api_key"))
	require.Equal(t, tool, query.Get("tool"))
	require.Equal(t, email, query.Get("email"))
}

func TestFetchArticle_NoIdentityParams(t *testing.T) {
	t.Parallel()
	query := fetchArticleQuery(t, Identity{})

	require.False(t, query.Has("api_key"))
	require.False(t, query.Has("tool"))
	require.False(t, query.Has("email"))
}
