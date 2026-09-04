package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testPMID   = "12345"
	testPMCID  = "PMC67890"
	testPDFURL = "https://bucket.example.com/PMC67890.1/PMC67890.1.pdf"
	efetchPath = "/efetch.fcgi"
)

// s3Path builds the deterministic PDF key suffix for a PMCID/version pair.
func s3Path(version int) string {
	return fmt.Sprintf("/%s.%d/%s.%d.pdf", testPMCID, version, testPMCID, version)
}

// mockEfetchHandler serves a minimal PubMed XML response for testPMID.
func mockEfetchHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(
		efetchPath,
		func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Query().Get("id") != testPMID {
				http.NotFound(writer, request)
				return
			}
			fmt.Fprintf(writer, `
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
		},
	)
	return mux
}

// mockS3Handler serves HEAD/GET at deterministic {PMCID}.{v} key paths.
// statusByVersion maps 1-based versions to HTTP statuses and bodies; missing
// entries return 404.
func mockS3Handler(
	statusByVersion map[int]int,
	body string,
) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			version := 0
			for v := 1; v <= maxPDFVersion; v++ {
				if request.URL.Path == s3Path(v) {
					version = v
					break
				}
			}
			if version == 0 {
				http.NotFound(writer, request)
				return
			}
			status, ok := statusByVersion[version]
			if !ok || status != http.StatusOK {
				if ok {
					writer.WriteHeader(status)
					return
				}
				http.NotFound(writer, request)
				return
			}
			if request.Method == http.MethodGet {
				fmt.Fprint(writer, body)
			}
		},
	)
}

// newTestPDFService wires a PDFService to mock efetch + mock S3 servers.
func newTestPDFService(
	s3handler http.Handler,
	s3body string,
) (*PDFService, *httptest.Server, *httptest.Server) {
	if s3handler == nil || s3body != "" {
		s3handler = mockS3Handler(mockOKVersions(1), s3body)
	}
	efetchServer := httptest.NewServer(mockEfetchHandler())
	s3Server := httptest.NewServer(s3handler)

	service := NewPDFService(WithHTTPClient(s3Server.Client()))
	service.pdfBaseURL = s3Server.URL
	service.articleService.baseURL = efetchServer.URL

	return service, efetchServer, s3Server
}

// mockOKVersions reports success at the given versions.
func mockOKVersions(versions ...int) map[int]int {
	result := make(map[int]int)
	for _, v := range versions {
		result[v] = http.StatusOK
	}
	return result
}

func TestNewPDFService(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service := NewPDFService()

	req.NotNil(service)
	req.NotNil(service.articleService)
	req.NotNil(service.httpClient)
	req.Equal(30*time.Second, service.httpClient.Timeout)
	req.Equal(
		"https://pmc-oa-opendata.s3.amazonaws.com",
		service.pdfBaseURL,
	)
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	customClient := &http.Client{Timeout: 10 * time.Second}
	service := NewPDFService(WithHTTPClient(customClient))

	req.Equal(customClient, service.httpClient)
	req.Equal(customClient, service.articleService.httpClient)
}

func TestResolvePDFURL_V1Success(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service, efetch, s3Server := newTestPDFService(nil, "")
	defer efetch.Close()
	defer s3Server.Close()

	url, err := service.resolvePDFURL(testPMCID)

	req.NoError(err)
	req.Equal(s3Server.URL+s3Path(1), url)
}

func TestResolvePDFURL_VersionFallback(t *testing.T) {
	t.Parallel()
	for _, miss := range []int{http.StatusForbidden, http.StatusNotFound} {
		name := fmt.Sprintf("after_%d", miss)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := require.New(t)
			handler := mockS3Handler(
				map[int]int{1: miss, 3: http.StatusOK},
				"",
			)
			service, efetch, s3Server := newTestPDFService(handler, "")
			defer efetch.Close()
			defer s3Server.Close()

			url, err := service.resolvePDFURL(testPMCID)

			req.NoError(err)
			req.Equal(s3Server.URL+s3Path(3), url)
		})
	}
}

func TestResolvePDFURL_ExhaustedIsNotFound(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	handler := mockS3Handler(nil, "")
	service, efetch, s3Server := newTestPDFService(handler, "")
	defer efetch.Close()
	defer s3Server.Close()

	_, err := service.resolvePDFURL(testPMCID)

	req.ErrorIs(err, errPDFNotFound)
}

func TestResolvePDFURL_UnexpectedStatus(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	handler := mockS3Handler(map[int]int{1: http.StatusInternalServerError}, "")
	service, efetch, s3Server := newTestPDFService(handler, "")
	defer efetch.Close()
	defer s3Server.Close()

	_, err := service.resolvePDFURL(testPMCID)

	req.Error(err)
	req.NotErrorIs(err, errPDFNotFound)
}

func TestIsPDFAvailable_ExhaustedReturnsFalse(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	handler := mockS3Handler(nil, "")
	service, efetch, s3Server := newTestPDFService(handler, "")
	defer efetch.Close()
	defer s3Server.Close()

	available, err := service.IsPDFAvailable(testPMID)

	req.NoError(err)
	req.False(available)
}

func TestStateManagement(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service, efetch, s3Server := newTestPDFService(nil, "")
	defer efetch.Close()
	defer s3Server.Close()

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

func TestFindPDFDownloadInfo_Success(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service, efetch, s3Server := newTestPDFService(nil, "")
	defer efetch.Close()
	defer s3Server.Close()

	info, err := service.findPDFDownloadInfo(testPMID)

	req.NoError(err)
	req.NotNil(info)
	req.Equal(testPMID, info.PMID)
	req.Equal(testPMCID, info.PMCID)
	req.Equal(s3Server.URL+s3Path(1), info.PDFURL)
}

func TestGetPDFURL_Success(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service, efetch, s3Server := newTestPDFService(nil, "")
	defer efetch.Close()
	defer s3Server.Close()

	// First, populate the state
	_, err := service.IsPDFAvailable(testPMID)
	req.NoError(err)

	pdfURL, err := service.GetPDFURL()
	req.NoError(err)
	req.Equal(s3Server.URL+s3Path(1), pdfURL)
}

func TestDownloadPDF_Success(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	service, efetch, s3Server := newTestPDFService(nil, "%PDF-1.4 fake")
	defer efetch.Close()
	defer s3Server.Close()

	available, err := service.IsPDFAvailable(testPMID)
	req.NoError(err)
	req.True(available)

	filePath := filepath.Join(t.TempDir(), "out.pdf")
	req.NoError(service.DownloadPDF(filePath))

	content, err := os.ReadFile(filePath)
	req.NoError(err)
	req.Equal("%PDF-1.4 fake", string(content))
}

func TestDownloadPDF_NoPartialFileOnFailure(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	handler := http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			if request.Method == http.MethodGet {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		},
	)
	service, efetch, s3Server := newTestPDFService(handler, "")
	defer efetch.Close()
	defer s3Server.Close()

	available, err := service.IsPDFAvailable(testPMID)
	req.NoError(err)
	req.True(available)

	filePath := filepath.Join(t.TempDir(), "out.pdf")
	err = service.DownloadPDF(filePath)
	req.Error(err)

	_, statErr := os.Stat(filePath)
	req.Error(statErr)
	req.True(os.IsNotExist(statErr))
}
