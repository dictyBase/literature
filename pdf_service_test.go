package literature

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testPMID   = "12345"
	testPMCID  = "PMC67890"
	testPDFURL = "ftp://ftp.ncbi.nlm.nih.gov/pub/pmc/oa_pdf/00/01/some.pdf"
	efetchPath = "/efetch.fcgi"
	oaPath     = "/oa.fcgi"
)

// mockAPIHandler returns a handler that serves mock XML responses for different endpoints.
func mockAPIHandler() http.Handler {
	mux := http.NewServeMux()

	// Mock for ArticleService efetch
	mux.HandleFunc(efetchPath, func(w http.ResponseWriter, r *http.Request) {
		pmid := r.URL.Query().Get("id")
		if pmid == testPMID {
			fmt.Fprintf(w, `
<PubmedArticleSet>
    <PubmedArticle>
        <MedlineCitation>
            <PMID>%s</PMID>
        </MedlineCitation>
        <PubmedData>
            <ArticleIdList>
                <ArticleId IdType="pmc">%s</ArticleId>
            </ArticleIdList>
        </PubmedData>
    </PubmedArticle>
</PubmedArticleSet>`, testPMID, testPMCID)
		} else {
			http.NotFound(w, r)
		}
	})

	// Mock for PDFService OA fetch
	mux.HandleFunc(oaPath, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == testPMCID {
			fmt.Fprintf(w, `
<OA>
    <records>
        <record id="%s">
            <link format="pdf" href="%s"/>
        </record>
    </records>
</OA>`, testPMCID, testPDFURL)
		} else {
			http.NotFound(w, r)
		}
	})

	return mux
}

// newTestPDFService creates a PDFService configured for testing with a mock server.
func newTestPDFService(handler http.Handler) (*PDFService, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := server.Client()

	service := NewPDFService(WithHTTPClient(client))
	service.oaBaseURL = server.URL
	// The internal article service also needs its URL updated.
	service.articleService.baseURL = server.URL

	return service, server
}

// TestNewPDFService from plan
func TestNewPDFService(t *testing.T) {
	req := require.New(t)
	service := NewPDFService()

	req.NotNil(service)
	req.NotNil(service.articleService)
	req.NotNil(service.httpClient)
	req.Equal(30*time.Second, service.httpClient.Timeout)
	req.Equal("https://www.ncbi.nlm.nih.gov/pmc/utils/oa", service.oaBaseURL)
}

// TestWithHTTPClient from plan
func TestWithHTTPClient(t *testing.T) {
	req := require.New(t)
	customClient := &http.Client{Timeout: 10 * time.Second}
	service := NewPDFService(WithHTTPClient(customClient))

	req.Equal(customClient, service.httpClient)
	req.Equal(customClient, service.articleService.httpClient)
}

// TestStateManagement covers clearState and GetCurrentPMID from plan
func TestStateManagement(t *testing.T) {
	req := require.New(t)
	service, server := newTestPDFService(mockAPIHandler())
	defer server.Close()

	// 1. Initial state
	req.Empty(service.GetCurrentPMID())
	req.Nil(service.downloadInfo)

	// 2. State after successful availability check
	available, err := service.IsPDFAvailable(testPMID)
	req.NoError(err)
	req.True(available)
	req.Equal(testPMID, service.GetCurrentPMID())
	req.NotNil(service.downloadInfo)

	// 3. State after clearState
	service.clearState()
	req.Empty(service.GetCurrentPMID())
	req.Nil(service.downloadInfo)
}

// TestIsPDFAvailable_Success from plan
func TestIsPDFAvailable_Success(t *testing.T) {
	req := require.New(t)
	service, server := newTestPDFService(mockAPIHandler())
	defer server.Close()

	// Set some dummy initial state to ensure it gets cleared
	service.currentPMID = "stale-pmid"
	service.downloadInfo = &PDFDownloadInfo{}

	available, err := service.IsPDFAvailable(testPMID)

	req.NoError(err)
	req.True(available)
	req.Equal(testPMID, service.currentPMID)
	req.NotNil(service.downloadInfo)
	req.Equal(testPMID, service.downloadInfo.PMID)
	req.Equal(testPMCID, service.downloadInfo.PMCID)
	req.Equal(testPDFURL, service.downloadInfo.PDFLink.HREF)
}

// TestFetchOADetails_Success from plan
func TestFetchOADetails_Success(t *testing.T) {
	req := require.New(t)
	service, server := newTestPDFService(mockAPIHandler())
	defer server.Close()

	oaRecord, err := service.fetchOADetails(testPMCID)

	req.NoError(err)
	req.NotNil(oaRecord)
	req.Equal(testPMCID, oaRecord.ID)
	req.Len(oaRecord.Links, 1)
	req.Equal("pdf", oaRecord.Links[0].Format)
	req.Equal(testPDFURL, oaRecord.Links[0].HREF)
}

// TestFindPDFDownloadInfo_Success from plan
func TestFindPDFDownloadInfo_Success(t *testing.T) {
	req := require.New(t)
	service, server := newTestPDFService(mockAPIHandler())
	defer server.Close()

	info, err := service.findPDFDownloadInfo(testPMID)

	req.NoError(err)
	req.NotNil(info)
	req.Equal(testPMID, info.PMID)
	req.Equal(testPMCID, info.PMCID)
	req.NotNil(info.PDFLink)
	req.Equal(testPDFURL, info.PDFLink.HREF)
}

// TestGetPDFURL_Success from plan
func TestGetPDFURL_Success(t *testing.T) {
	req := require.New(t)
	service, server := newTestPDFService(mockAPIHandler())
	defer server.Close()

	// First, populate the state
	_, err := service.IsPDFAvailable(testPMID)
	req.NoError(err)

	// Then, test GetPDFURL
	pdfURL, err := service.GetPDFURL()
	req.NoError(err)
	req.Equal(testPDFURL, pdfURL)
}

// TestIsPMCID from plan
func TestIsPMCID(t *testing.T) {
	testCases := []struct {
		name     string
		id       ArticleID
		expected bool
	}{
		{"Valid PMCID", ArticleID{IDType: "pmc", Value: "PMC123"}, true},
		{"Not PMCID", ArticleID{IDType: "doi", Value: "10.1000/xyz"}, false},
		{"Empty IDType", ArticleID{Value: "123"}, false},
		{
			"Case sensitive check",
			ArticleID{IDType: "Pmc", Value: "PMC123"},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := require.New(t)
			result := IsPMCID(tc.id)
			req.Equal(tc.expected, result)
		})
	}
}

// TestIsPDFLink from plan
func TestIsPDFLink(t *testing.T) {
	testCases := []struct {
		name     string
		link     OALink
		expected bool
	}{
		{"Valid PDF link", OALink{Format: "pdf", HREF: "some_url"}, true},
		{"Not a PDF link", OALink{Format: "tgz", HREF: "some_url"}, false},
		{"Empty format", OALink{HREF: "some_url"}, false},
		{
			"Case sensitive check",
			OALink{Format: "PDF", HREF: "some_url"},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := require.New(t)
			result := IsPDFLink(tc.link)
			req.Equal(tc.expected, result)
		})
	}
}
