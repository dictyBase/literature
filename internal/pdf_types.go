package internal

import (
	"encoding/xml"
	"fmt"
)

// OAResponse represents the top-level XML structure for the OA FTP service.
type OAResponse struct {
	XMLName xml.Name `xml:"OA"`
	Records struct {
		Records []OARecord `xml:"record"`
	} `xml:"records"`
}

// OARecord contains information about a specific record, including download links.
type OARecord struct {
	XMLName xml.Name `xml:"record"`
	ID      string   `xml:"id,attr"`
	Links   []OALink `xml:"link"`
}

// OALink represents a download link for a specific format (e.g., pdf, tgz).
type OALink struct {
	XMLName xml.Name `xml:"link"`
	Format  string   `xml:"format,attr"`
	HREF    string   `xml:"href,attr"`
}

// PDFDownloadInfo contains information needed to download a PDF.
type PDFDownloadInfo struct {
	PMID    string
	PMCID   string
	PDFLink *OALink
}

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
