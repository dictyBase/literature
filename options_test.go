package literature

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	identityTestAPIKey = "public-option-api-key"
	identityTestTool   = "literature-integration-test"
	identityTestEmail  = "api-test@example.org"
)

type identityOptionsRequest struct {
	path      string
	query     url.Values
	userAgent string
}

func newIdentityOptionsClient(
	t *testing.T,
	handler http.Handler,
	userAgent string,
) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := New(
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithUserAgent(userAgent),
		WithAPIKey(identityTestAPIKey),
		WithTool(identityTestTool),
		WithEmail(identityTestEmail),
	)
	require.NoError(t, err)
	return client, server
}

func captureIdentityOptionsRequest(
	requests chan<- identityOptionsRequest,
	body string,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		requests <- identityOptionsRequest{
			path:      request.URL.Path,
			query:     request.URL.Query(),
			userAgent: request.Header.Get("User-Agent"),
		}
		_, _ = fmt.Fprint(writer, body)
	}
}

func requireIdentityOptionsRequest(
	t *testing.T,
	request identityOptionsRequest,
	userAgent string,
) {
	t.Helper()
	req := require.New(t)
	req.Equal(identityTestAPIKey, request.query.Get("api_key"))
	req.Equal(identityTestTool, request.query.Get("tool"))
	req.Equal(identityTestEmail, request.query.Get("email"))
	req.Equal(userAgent, request.userAgent)
}

func TestWithTool_RejectsSpaces(t *testing.T) {
	t.Parallel()
	_, err := New(WithTool("tool with spaces"))

	req := require.New(t)
	req.Error(err)
	var libraryError *Error
	req.ErrorAs(err, &libraryError)
	req.Equal(ErrorTypeInvalidInput, libraryError.Type)
}

func TestWithEmail_RejectsMalformed(t *testing.T) {
	t.Parallel()
	_, err := New(WithEmail("no-at-sign"))

	req := require.New(t)
	req.Error(err)
	var libraryError *Error
	req.ErrorAs(err, &libraryError)
	req.Equal(ErrorTypeInvalidInput, libraryError.Type)
}

func TestIdentityOptions_EmptyValuesAccepted(t *testing.T) {
	t.Parallel()
	_, err := New(
		WithAPIKey(""),
		WithTool(""),
		WithEmail(""),
	)

	require.NoError(t, err)
}

func TestNew_IdentityParamsOnArticleRequest(t *testing.T) {
	t.Parallel()
	const userAgent = "literature-root-test/1.0"
	requests := make(chan identityOptionsRequest, 1)
	handler := captureIdentityOptionsRequest(requests, `
<PubmedArticleSet>
	<PubmedArticle>
		<MedlineCitation><PMID>123</PMID></MedlineCitation>
	</PubmedArticle>
</PubmedArticleSet>`)
	client, server := newIdentityOptionsClient(t, handler, userAgent)
	defer server.Close()

	_, err := client.GetArticle("123")
	require.NoError(t, err)
	requireIdentityOptionsRequest(t, <-requests, userAgent)
}

func TestNew_IdentityParamsOnSearchRequests(t *testing.T) {
	t.Parallel()
	const userAgent = "literature-root-search-test/1.0"
	requests := make(chan identityOptionsRequest, 2)
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/esearch.fcgi",
		captureIdentityOptionsRequest(requests, "<eSearchResult></eSearchResult>"),
	)
	mux.HandleFunc(
		"/efetch.fcgi",
		captureIdentityOptionsRequest(requests, "<PubmedArticleSet></PubmedArticleSet>"),
	)
	client, server := newIdentityOptionsClient(t, mux, userAgent)
	defer server.Close()

	req := require.New(t)
	_, err := client.searchService.SearchPubMed("cancer & therapy", 7, 3)
	req.NoError(err)
	_, err = client.searchService.FetchPubMedDetails("web env", "2", 7, 3)
	req.NoError(err)

	esearchRequest := <-requests
	req.Equal("/esearch.fcgi", esearchRequest.path)
	requireIdentityOptionsRequest(t, esearchRequest, userAgent)
	req.Equal("cancer & therapy", esearchRequest.query.Get("term"))
	req.Equal("7", esearchRequest.query.Get("retmax"))
	req.Equal("3", esearchRequest.query.Get("retstart"))
	req.Equal("y", esearchRequest.query.Get("usehistory"))

	efetchRequest := <-requests
	req.Equal("/efetch.fcgi", efetchRequest.path)
	requireIdentityOptionsRequest(t, efetchRequest, userAgent)
	req.Equal("web env", efetchRequest.query.Get("WebEnv"))
	req.Equal("2", efetchRequest.query.Get("query_key"))
	req.Equal("7", efetchRequest.query.Get("retmax"))
	req.Equal("3", efetchRequest.query.Get("retstart"))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func newTestHTTPResponse(request *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}

func TestNew_IdentityParamsOnPDFRequest(t *testing.T) {
	t.Parallel()
	requests := make(chan url.Values, 1)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Host {
		case "ncbi.test":
			requests <- request.URL.Query()
			return newTestHTTPResponse(request, `
<PubmedArticleSet>
	<PubmedArticle>
		<MedlineCitation><PMID>123</PMID></MedlineCitation>
		<PubmedData><ArticleIdList>
			<ArticleId IdType="pmc">PMC123</ArticleId>
		</ArticleIdList></PubmedData>
	</PubmedArticle>
</PubmedArticleSet>`), nil
		case "pmc-oa-opendata.s3.amazonaws.com":
			return newTestHTTPResponse(request, ""), nil
		default:
			return nil, fmt.Errorf("unexpected request host %q", request.URL.Host)
		}
	})
	client, err := New(
		WithHTTPClient(&http.Client{Transport: transport}),
		WithBaseURL("https://ncbi.test"),
		WithAPIKey(identityTestAPIKey),
		WithTool(identityTestTool),
		WithEmail(identityTestEmail),
	)
	req := require.New(t)
	req.NoError(err)
	available, err := client.HasPDF("123")
	req.NoError(err)
	req.True(available)

	query := <-requests
	req.Equal(identityTestAPIKey, query.Get("api_key"))
	req.Equal(identityTestTool, query.Get("tool"))
	req.Equal(identityTestEmail, query.Get("email"))
}
