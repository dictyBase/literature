package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// PDFService handles PDF link discovery and downloading.
type PDFService struct {
	articleService *ArticleService
	httpClient     *http.Client
	oaBaseURL      string
	// State for caching download readiness
	currentPMID  string
	downloadInfo *PDFDownloadInfo
}

// PDFServiceOption configures PDFService behavior.
type PDFServiceOption func(*PDFService)

// WithHTTPClient sets a custom HTTP client for the PDF service.
func WithHTTPClient(client *http.Client) PDFServiceOption {
	return func(s *PDFService) {
		s.httpClient = client
		s.articleService.httpClient = client
	}
}

// NewPDFService creates a new PDFService with the given options.
func NewPDFService(options ...PDFServiceOption) *PDFService {
	service := &PDFService{
		articleService: NewArticleService(),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		oaBaseURL:      "https://www.ncbi.nlm.nih.gov/pmc/utils/oa",
	}

	for _, option := range options {
		option(service)
	}

	return service
}

// IsPDFAvailable checks if a PDF is available for the given PMID and caches download info.
// Must be called before DownloadPDF.
func (s *PDFService) IsPDFAvailable(pmid string) (bool, error) {
	s.clearState() // Clear any previous state

	info, err := s.findPDFDownloadInfo(pmid)
	if err != nil {
		var pdfErr *PDFError
		if errors.As(err, &pdfErr) {
			switch pdfErr.Type {
			case PDFErrorPMCIDNotFound, PDFErrorPDFNotAvailable:
				return false, nil
			}
		}
		return false, err
	}

	// Cache the download info for subsequent DownloadPDF call
	s.currentPMID = pmid
	s.downloadInfo = info
	return true, nil
}

// clearState clears the cached download state.
func (s *PDFService) clearState() {
	s.currentPMID = ""
	s.downloadInfo = nil
}

// GetCurrentPMID returns the currently cached PMID, empty string if none.
func (s *PDFService) GetCurrentPMID() string {
	return s.currentPMID
}

// fetchOADetails retrieves Open Access details for a given PMC ID.
func (s *PDFService) fetchOADetails(pmcID string) (*OARecord, error) {
	oaURL := fmt.Sprintf(
		"%s/oa.fcgi?id=%s",
		s.oaBaseURL,
		pmcID,
	)

	// #nosec G107
	resp, err := s.httpClient.Get(oaURL)
	if err != nil {
		return nil, fmt.Errorf("error making OA request: %w", err)
	}
	defer resp.Body.Close()

	oaResponse := &OAResponse{}
	if err := xml.NewDecoder(resp.Body).Decode(oaResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling OA response: %w", err)
	}
	if len(oaResponse.Records.Records) == 0 {
		return nil, fmt.Errorf("no records found in OA response for %s", pmcID)
	}
	// Assuming the first record is the one we want.
	return &oaResponse.Records.Records[0], nil
}

// FindPDFDownloadInfo locates PDF download information for the given PMID.
func (s *PDFService) findPDFDownloadInfo(
	pmid string,
) (*PDFDownloadInfo, error) {
	article, err := s.articleService.FetchArticle(pmid)
	if err != nil {
		return nil, err
	}

	pmcArticleID, found := Find(
		article.PubmedData.ArticleIdList.ArticleIDs,
		IsPMCID,
	)
	if !found {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorPMCIDNotFound,
		}
	}

	oaRecord, err := s.fetchOADetails(pmcArticleID.Value)
	if err != nil {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorPDFNotAvailable,
			Err:  err,
		}
	}

	pdfLink, found := Find(oaRecord.Links, IsPDFLink)
	if !found {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorPDFNotAvailable,
		}
	}

	return &PDFDownloadInfo{
		PMID:    pmid,
		PMCID:   pmcArticleID.Value,
		PDFLink: pdfLink,
	}, nil
}

// DownloadPDF downloads the PDF using cached download info to the specified file.
// IsPDFAvailable must be called first and return true.
func (s *PDFService) DownloadPDF(filePath string) error {
	if s.downloadInfo == nil {
		return &PDFError{
			PMID: s.currentPMID,
			Type: PDFErrorDownloadFailed,
			Err:  fmt.Errorf("must call IsPDFAvailable first and confirm PDF is available"),
		}
	}

	// Clear state after download regardless of success/failure
	defer s.clearState()

	err := s.downloadFromFTP(s.downloadInfo.PDFLink.HREF, filePath)
	if err != nil {
		return &PDFError{
			PMID: s.downloadInfo.PMID,
			Type: PDFErrorDownloadFailed,
			Err:  err,
		}
	}

	return nil
}

// downloadFromFTP downloads a file from an FTP URL to the specified file path.
func (s *PDFService) downloadFromFTP(ftpURL, filePath string) error {
	parsedURL, err := url.Parse(ftpURL)
	if err != nil {
		return fmt.Errorf("invalid FTP URL: %w", err)
	}

	host := parsedURL.Host
	if !strings.Contains(host, ":") {
		host += ":21" // Default FTP port
	}

	fclient, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}
	defer func() {
		_ = fclient.Quit() // Ignore quit errors
	}()

	err = fclient.Login("anonymous", "anonymous")
	if err != nil {
		return fmt.Errorf("FTP login failed: %w", err)
	}

	path := parsedURL.Path

	res, err := fclient.Retr(path)
	if err != nil {
		return fmt.Errorf("failed to retrieve file from FTP: %w", err)
	}
	defer res.Close()

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetPDFURL returns the direct download URL using cached download info.
// IsPDFAvailable must be called first and return true.
func (s *PDFService) GetPDFURL() (string, error) {
	if s.downloadInfo == nil {
		return "", &PDFError{
			PMID: s.currentPMID,
			Type: PDFErrorDownloadFailed,
			Err:  fmt.Errorf("must call IsPDFAvailable first and confirm PDF is available"),
		}
	}

	return s.downloadInfo.PDFLink.HREF, nil
}

// DownloadArticlePDF is a convenience method that combines availability check and downloading.
func (s *PDFService) DownloadArticlePDF(pmid, filePath string) error {
	available, err := s.IsPDFAvailable(pmid)
	if err != nil {
		return err
	}

	if !available {
		return &PDFError{
			PMID: pmid,
			Type: PDFErrorPDFNotAvailable,
		}
	}

	return s.DownloadPDF(filePath)
}
