package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestArticleService(
	handler http.HandlerFunc,
) (*ArticleService, *httptest.Server) {
	server := httptest.NewServer(handler)
	service := &ArticleService{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}
	return service, server
}

func TestNewArticleService(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service := NewArticleService()

	req.NotNil(service)
	req.Equal(30*time.Second, service.httpClient.Timeout)

	expectedBaseURL := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils"
	req.Equal(expectedBaseURL, service.baseURL)
}

func TestFetchArticle_Success(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	const pmid = "12345678"
	const doi = "10.1234/test.doi"
	const title = "A Test Article"

	xmlResponse := fmt.Sprintf(`
<PubmedArticleSet>
    <PubmedArticle>
        <MedlineCitation>
            <PMID>%s</PMID>
            <Article>
                <ArticleTitle>%s</ArticleTitle>
            </Article>
        </MedlineCitation>
        <PubmedData>
            <ArticleIdList>
                <ArticleId IdType="doi">%s</ArticleId>
            </ArticleIdList>
        </PubmedData>
    </PubmedArticle>
</PubmedArticleSet>`, pmid, title, doi)

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := fmt.Fprint(writer, xmlResponse)
		require.NoError(t, err)
	})
	service, server := newTestArticleService(handler)
	defer server.Close()

	article, err := service.FetchArticle(pmid)
	req.NoError(err)
	req.NotNil(article)
	req.Equal(pmid, article.GetPMID())
	req.Equal(title, article.GetTitle())
}

func TestFetchArticle_HTTPErrors(t *testing.T) {
	t.Parallel()
	t.Run("request fails", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)
		service, server := newTestArticleService(
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}),
		)
		server.Close() // Close immediately to simulate network failure

		pmid := "123"
		_, err := service.FetchArticle(pmid)

		req.Error(err)

		var pdfErr *PDFError
		req.ErrorAs(err, &pdfErr)

		req.Equal(PDFErrorArticleNotFound, pdfErr.Type)
		req.Equal(pmid, pdfErr.PMID)
	})

	t.Run("timeout", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)
		service, server := newTestArticleService(
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				time.Sleep(100 * time.Millisecond)
			}),
		)
		defer server.Close()

		// Replace client with one that has a short timeout
		service.httpClient = &http.Client{Timeout: 50 * time.Millisecond}

		pmid := "timeout"
		_, err := service.FetchArticle(pmid)

		req.Error(err)

		var pdfErr *PDFError
		req.ErrorAs(err, &pdfErr)

		isTimeoutError := strings.Contains(
			err.Error(),
			"context deadline exceeded",
		) ||
			strings.Contains(err.Error(), "Client.Timeout exceeded")
		req.True(isTimeoutError, "expected timeout error, but got: %v", err)
	})
}

func TestFetchArticle_XMLParsingErrors(t *testing.T) {
	t.Parallel()
	t.Run("invalid XML response", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)
		handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, err := fmt.Fprint(writer, "<malformed xml")
			require.NoError(t, err)
		})
		service, server := newTestArticleService(handler)
		defer server.Close()

		pmid := "456"
		_, err := service.FetchArticle(pmid)

		req.Error(err)

		var pdfErr *PDFError
		req.ErrorAs(err, &pdfErr)

		req.Equal(PDFErrorArticleNotFound, pdfErr.Type)
		req.ErrorContains(err, "error unmarshaling efetch XML")
	})

	t.Run("empty response", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)
		service, server := newTestArticleService(
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Return empty body
			}),
		)
		defer server.Close()

		pmid := "789"
		_, err := service.FetchArticle(pmid)

		req.Error(err)

		var pdfErr *PDFError
		req.ErrorAs(err, &pdfErr)

		req.ErrorContains(pdfErr.Unwrap(), "EOF")
	})
}

func TestFetchArticle_BusinessLogicErrors(t *testing.T) {
	t.Parallel()
	t.Run("no articles found", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)
		handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, err := fmt.Fprint(writer, "<PubmedArticleSet></PubmedArticleSet>")
			require.NoError(t, err)
		})
		service, server := newTestArticleService(handler)
		defer server.Close()

		pmid := "101112"
		_, err := service.FetchArticle(pmid)

		req.Error(err)

		var pdfErr *PDFError
		req.ErrorAs(err, &pdfErr)

		req.Equal(PDFErrorArticleNotFound, pdfErr.Type)
		req.ErrorContains(err, "no articles found")
	})

	t.Run("empty pmid", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)
		handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, err := fmt.Fprint(writer, "<PubmedArticleSet></PubmedArticleSet>")
			require.NoError(t, err)
		})
		service, server := newTestArticleService(handler)
		defer server.Close()

		_, err := service.FetchArticle("")

		req.Error(err)

		var pdfErr *PDFError
		req.ErrorAs(err, &pdfErr)

		req.ErrorContains(err, "no articles found")
	})
}

func TestFormatAuthor(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		author   Author
		expected string
	}{
		{
			name:     "ValidAuthor",
			author:   Author{ForeName: "John", LastName: "Doe"},
			expected: "John Doe",
		},
		{
			name:     "EmptyForeName",
			author:   Author{LastName: "Doe"},
			expected: " Doe",
		},
		{
			name:     "EmptyLastName",
			author:   Author{ForeName: "John"},
			expected: "John ",
		},
		{
			name:     "BothEmpty",
			author:   Author{},
			expected: " ",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			req := require.New(t)
			result := formatAuthor(testCase.author)
			req.Equal(testCase.expected, result)
		})
	}
}

func TestIsDOI(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		id       ArticleID
		expected bool
	}{
		{
			name:     "ValidDOI",
			id:       ArticleID{IDType: "doi", Value: "10.1000/xyz123"},
			expected: true,
		},
		{
			name:     "NotDOI",
			id:       ArticleID{IDType: "pmc", Value: "PMC12345"},
			expected: false,
		},
		{
			name:     "EmptyIDType",
			id:       ArticleID{Value: "12345"},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			req := require.New(t)
			result := IsDOI(testCase.id)
			req.Equal(testCase.expected, result)
		})
	}
}
