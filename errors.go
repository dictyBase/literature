package main

import "fmt"

// PDFErrorType represents different types of PDF download errors.
type PDFErrorType int

const (
	PDFErrorArticleNotFound PDFErrorType = iota
	PDFErrorPMCIDNotFound
	PDFErrorPDFNotAvailable
	PDFErrorDownloadFailed
)

// PDFError provides detailed error information for PDF download failures.
type PDFError struct {
	PMID string
	Type PDFErrorType
	Err  error
}

func (e *PDFError) Error() string {
	switch e.Type {
	case PDFErrorArticleNotFound:
		return fmt.Sprintf("article not found for PMID %s: %v", e.PMID, e.Err)
	case PDFErrorPMCIDNotFound:
		return fmt.Sprintf("no PMC ID found for PMID %s", e.PMID)
	case PDFErrorPDFNotAvailable:
		return fmt.Sprintf("PDF not available for PMID %s", e.PMID)
	case PDFErrorDownloadFailed:
		return fmt.Sprintf("download failed for PMID %s: %v", e.PMID, e.Err)
	default:
		return fmt.Sprintf("unknown error for PMID %s: %v", e.PMID, e.Err)
	}
}

func (e *PDFError) Unwrap() error {
	return e.Err
}
