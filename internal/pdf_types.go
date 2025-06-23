package literature

import "encoding/xml"

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
