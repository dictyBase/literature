package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// maxPDFVersion bounds the version probing loop. PMC article versions rarely
// exceed a handful, so a small fixed bound keeps lookups cheap.
const maxPDFVersion = 10

// errPDFNotFound marks a version-exhaustion miss so callers can distinguish
// "no PDF for this PMCID" from transport or unexpected status failures.
var errPDFNotFound = errors.New("no PDF found within probed versions")

// PDFService handles PDF link discovery and downloading.
type PDFService struct {
	articleService *ArticleService
	httpClient     *http.Client
	pdfBaseURL     string
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
		pdfBaseURL:     "https://pmc-oa-opendata.s3.amazonaws.com",
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
		if pdfErr, ok := errors.AsType[*PDFError](err); ok {
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

// resolvePDFURL probes the PMC Open Access S3 bucket for
// {PMCID}.{version}/{PMCID}.{version}.pdf, versions 1..maxPDFVersion.
// Both 404 and 403 count as misses: S3 answers HEAD with 403 for absent keys
// when anonymous callers lack ListBucket permission.
func (s *PDFService) resolvePDFURL(pmcid string) (string, error) {
	var lastErr error

	for version := 1; version <= maxPDFVersion; version++ {
		candidate := fmt.Sprintf(
			"%s/%s.%d/%s.%d.pdf",
			s.pdfBaseURL,
			pmcid,
			version,
			pmcid,
			version,
		)
		// #nosec G107
		resp, err := s.httpClient.Head(candidate)
		if err != nil {
			lastErr = fmt.Errorf("HEAD %s: %w", candidate, err)
			return "", lastErr
		}
		status := resp.StatusCode
		_ = resp.Body.Close()

		switch {
		case status >= 200 && status < 300:
			return candidate, nil
		case status == 404 || status == 403:
			continue
		default:
			return "", fmt.Errorf(
				"unexpected status %d probing %s",
				status,
				candidate,
			)
		}
	}

	return "", errPDFNotFound
}

// findPDFDownloadInfo locates PDF download information for the given PMID.
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

	pdfURL, err := s.resolvePDFURL(pmcArticleID.Value)
	if err != nil {
		if errors.Is(err, errPDFNotFound) {
			return nil, &PDFError{
				PMID: pmid,
				Type: PDFErrorPDFNotAvailable,
			}
		}
		return nil, err
	}

	return &PDFDownloadInfo{
		PMID:   pmid,
		PMCID:  pmcArticleID.Value,
		PDFURL: pdfURL,
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

	err := s.downloadFromHTTP(s.downloadInfo.PDFURL, filePath)
	if err != nil {
		return &PDFError{
			PMID: s.downloadInfo.PMID,
			Type: PDFErrorDownloadFailed,
			Err:  err,
		}
	}

	return nil
}

// downloadFromHTTP downloads a file over HTTP(S) to the given path. Any
// partial output file is removed when the fetch or copy fails.
func (s *PDFService) downloadFromHTTP(url, filePath string) error {
	// #nosec G107
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"unexpected status %d downloading PDF from %s",
			resp.StatusCode,
			url,
		)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, resp.Body); err != nil {
		_ = os.Remove(filePath)
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

	return s.downloadInfo.PDFURL, nil
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
