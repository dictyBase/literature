package literature

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dictybase/literature/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testParams holds common test parameters.
type testParams struct {
	t      *testing.T
	client *EuropePMCClient
	assert *require.Assertions
}

// createTestEuropePMCClient creates a test client with a mock server.
func createTestEuropePMCClient(
	t *testing.T,
	handler http.HandlerFunc,
) (*EuropePMCClient, *httptest.Server) {
	t.Helper()

	// Create test server with the handler
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Route all requests to our test handler
		handler(responseWriter, request)
	}))

	// Create client configured to use the test server
	client, err := NewEuropePMCClient(
		WithEuropePMCHTTPClient(server.Client()),
		WithEuropePMCBaseURL(server.URL),
	)
	require.NoError(t, err)

	return client, server
}

// mockEuropePMCSuccessResponse creates a mock EuropePMC API response.
func mockEuropePMCSuccessResponse(pmid, doi, title string) *internal.EuropePMCAPIResponse {
	return &internal.EuropePMCAPIResponse{
		Version:  "6.8",
		HitCount: 1,
		ResultList: internal.EuropePMCAPIResultList{
			Result: []internal.EuropePMCAPIArticle{
				{
					ID:           pmid,
					Source:       "MED",
					PMID:         pmid,
					DOI:          doi,
					Title:        title,
					AuthorString: "Smith J, Doe A",
					AbstractText: "This is a test abstract",
					PubYear:      "2023",
					IsOpenAccess: "N",
					HasPDF:       "Y",
					Language:     "eng",
					JournalInfo: internal.EuropePMCAPIJournalInfo{
						Journal: internal.EuropePMCAPIJournal{
							Title: "Test Journal",
							ISSN:  "1234-5678",
						},
						Volume: "10",
						Issue:  "1",
					},
					AuthorList: internal.EuropePMCAPIAuthorList{
						Author: []internal.EuropePMCAPIAuthor{
							{
								FullName:  "John Smith",
								FirstName: "John",
								LastName:  "Smith",
								Initials:  "J",
							},
						},
					},
					PubTypeList: internal.EuropePMCPubTypeList{
						PubType: []string{"Journal Article"},
					},
					KeywordList: internal.EuropePMCKeywordList{
						Keyword: []string{"test", "example"},
					},
					FullTextURLList: internal.EuropePMCFullTextURLList{
						FullTextURL: []internal.EuropePMCAPIFullTextURL{
							{
								Availability:     "Open access",
								AvailabilityCode: "OA",
								DocumentStyle:    "pdf",
								Site:             "Europe PMC",
								URL:              "https://example.com/pdf",
							},
						},
					},
				},
			},
		},
	}
}

// mockEuropePMCEmptyResponse creates a mock empty EuropePMC API response.
func mockEuropePMCEmptyResponse() *internal.EuropePMCAPIResponse {
	return &internal.EuropePMCAPIResponse{
		Version:    "6.8",
		HitCount:   0,
		ResultList: internal.EuropePMCAPIResultList{Result: []internal.EuropePMCAPIArticle{}},
	}
}

func TestGetArticleByDOI_Success(t *testing.T) {
	t.Parallel()
	const testDOI = "10.1038/nature12373"
	const testPMID = "23842501"
	const testTitle = "A draft sequence of the Neandertal genome"

	handler := func(responseWriter http.ResponseWriter, request *http.Request) {
		// Verify the query parameter contains DOI format
		query := request.URL.Query().Get("query")
		expectedQuery := fmt.Sprintf("DOI:\"%s\"", testDOI)
		assert.Contains(t, query, expectedQuery, "Query should contain DOI format")

		mockResponse := mockEuropePMCSuccessResponse(testPMID, testDOI, testTitle)
		responseWriter.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(responseWriter).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	client, server := createTestEuropePMCClient(t, handler)
	defer server.Close()

	// TODO(human): Implement the core test logic here
	// Test the GetArticleByDOI method and verify the returned article
	// You should assert that:
	// 1. No error is returned
	// 2. The article is not nil
	// 3. The article's DOI matches the input DOI
	// 4. The article's title matches the expected title
	// 5. Other key fields are properly populated

	params := &testParams{
		t:      t,
		client: client,
		assert: require.New(t),
	}
	testGetArticleByDOICore(params, testDOI, testTitle)
}

func TestGetArticleByDOI_NotFound(t *testing.T) {
	t.Parallel()
	const testDOI = "10.1234/nonexistent.doi"

	handler := func(responseWriter http.ResponseWriter, request *http.Request) {
		mockResponse := mockEuropePMCEmptyResponse()
		responseWriter.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(responseWriter).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	client, server := createTestEuropePMCClient(t, handler)
	defer server.Close()

	params := &testParams{
		t:      t,
		client: client,
		assert: require.New(t),
	}
	testGetArticleByDOINotFound(params, testDOI)
}

func TestGetArticleByDOI_InvalidInput(t *testing.T) {
	t.Parallel()
	client, err := NewEuropePMCClient()
	require.NoError(t, err)

	params := &testParams{
		t:      t,
		client: client,
		assert: require.New(t),
	}
	testGetArticleByDOIInvalidInput(params)
}

func TestGetArticleByDOI_NetworkError(t *testing.T) {
	t.Parallel()
	const testDOI = "10.1038/nature12373"

	handler := func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	}

	client, server := createTestEuropePMCClient(t, handler)
	defer server.Close()

	params := &testParams{
		t:      t,
		client: client,
		assert: require.New(t),
	}
	testGetArticleByDOINetworkError(params, testDOI)
}

// Test helper functions

func testGetArticleByDOICore(params *testParams, testDOI, expectedTitle string) {
	params.t.Helper()

	article, err := params.client.GetArticleByDOI(testDOI)

	// Assert no error occurred
	params.assert.NoError(err, "should successfully fetch article by DOI")

	// Assert article is not nil
	params.assert.NotNil(article, "article should not be nil")

	// Assert DOI matches input
	params.assert.Equal(testDOI, article.DOI, "article DOI should match input DOI")

	// Assert title matches expected value
	params.assert.Equal(expectedTitle, article.Title, "article title should match expected title")

	// Assert other key fields are properly populated
	params.assert.Equal("23842501", article.PMID, "article PMID should be correctly set")
	params.assert.Equal("Smith J, Doe A", article.AuthorString, "article author string should be correctly set")
	params.assert.True(article.HasPDF, "article should indicate PDF availability")
	params.assert.False(article.IsOpenAccess, "article should indicate correct open access status")
	params.assert.Equal("eng", article.Language, "article language should be correctly set")
	params.assert.Equal("2023", article.PubYear, "article publication year should be correctly set")
	params.assert.Equal("This is a test abstract", article.Abstract, "article abstract should be correctly set")

	// Assert journal information is populated
	params.assert.NotNil(article.Journal, "journal information should not be nil")
	params.assert.Equal("Test Journal", article.Journal.Title, "journal title should be correctly set")
	params.assert.Equal("1234-5678", article.Journal.ISSN, "journal ISSN should be correctly set")
	params.assert.Equal("10", article.Journal.Volume, "journal volume should be correctly set")
	params.assert.Equal("1", article.Journal.Issue, "journal issue should be correctly set")

	// Assert authors are populated
	params.assert.NotEmpty(article.Authors, "authors list should not be empty")
	params.assert.Equal("John Smith", article.Authors[0].FullName, "first author full name should be correctly set")
	params.assert.Equal("John", article.Authors[0].FirstName, "first author first name should be correctly set")
	params.assert.Equal("Smith", article.Authors[0].LastName, "first author last name should be correctly set")
	params.assert.Equal("J", article.Authors[0].Initials, "first author initials should be correctly set")

	// Assert publication types are populated
	params.assert.NotEmpty(article.PubTypes, "publication types should not be empty")
	params.assert.Contains(article.PubTypes, "Journal Article", "should contain expected publication type")

	// Assert keywords are populated
	params.assert.NotEmpty(article.Keywords, "keywords should not be empty")
	params.assert.Contains(article.Keywords, "test", "should contain expected keyword 'test'")
	params.assert.Contains(article.Keywords, "example", "should contain expected keyword 'example'")

	// Assert full text URLs are populated
	params.assert.NotEmpty(article.FullTextURLs, "full text URLs should not be empty")
	params.assert.Equal("pdf", article.FullTextURLs[0].DocumentStyle, "document style should be PDF")
	params.assert.Equal("Europe PMC", article.FullTextURLs[0].Site, "site should be correctly set")
	params.assert.Equal("https://example.com/pdf", article.FullTextURLs[0].URL, "URL should be correctly set")
}

func testGetArticleByDOINotFound(params *testParams, testDOI string) {
	params.t.Helper()
	article, err := params.client.GetArticleByDOI(testDOI)

	params.assert.Error(err, "should return error for non-existent DOI")
	params.assert.Nil(article, "article should be nil for non-existent DOI")

	var litErr *Error
	params.assert.ErrorAs(err, &litErr, "error should be of type *Error")
	params.assert.Equal(ErrorTypeArticleNotFound, litErr.Type, "error type should be article not found")
	params.assert.Equal(testDOI, litErr.DOI, "error should contain the DOI")
}

func testGetArticleByDOIInvalidInput(params *testParams) {
	params.t.Helper()
	// Test empty DOI
	article, err := params.client.GetArticleByDOI("")

	params.assert.Error(err, "should return error for empty DOI")
	params.assert.Nil(article, "article should be nil for empty DOI")

	var litErr *Error
	params.assert.ErrorAs(err, &litErr, "error should be of type *Error")
	params.assert.Equal(ErrorTypeInvalidInput, litErr.Type, "error type should be invalid input")
}

func testGetArticleByDOINetworkError(params *testParams, testDOI string) {
	params.t.Helper()
	article, err := params.client.GetArticleByDOI(testDOI)

	params.assert.Error(err, "should return error for network failure")
	params.assert.Nil(article, "article should be nil for network failure")

	var litErr *Error
	params.assert.ErrorAs(err, &litErr, "error should be of type *Error")
	params.assert.Equal(ErrorTypeAPIError, litErr.Type, "error type should be API error")
	params.assert.Equal(testDOI, litErr.DOI, "error should contain the DOI")
}
