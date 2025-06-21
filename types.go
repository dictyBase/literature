package main

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

// ESearchResult represents the top-level XML structure for esearch.
type ESearchResult struct {
	XMLName  xml.Name `xml:"eSearchResult"`
	Count    string   `xml:"Count"`
	RetMax   string   `xml:"RetMax"`
	RetStart string   `xml:"RetStart"`
	QueryKey string   `xml:"QueryKey"`
	WebEnv   string   `xml:"WebEnv"`
	IDList   struct {
		IDs []string `xml:"Id"`
	} `xml:"IdList"`
}

// PubMedArticleSet represents the top-level XML structure for efetch.
type PubMedArticleSet struct {
	XMLName        xml.Name        `xml:"PubmedArticleSet"`
	PubMedArticles []PubMedArticle `xml:"PubmedArticle"`
}

// Author represents an author of an article.
type Author struct {
	LastName string `xml:"LastName"`
	ForeName string `xml:"ForeName"`
}

// ArticleID represents an identifier for an article, like DOI or PMID.
type ArticleID struct {
	IDType string `xml:"IdType,attr"`
	Value  string `xml:",chardata"`
}

// PubMedArticle represents a single PubMed article.
type PubMedArticle struct {
	XMLName         xml.Name `xml:"PubmedArticle"`
	MedlineCitation struct {
		Article struct {
			Journal struct {
				Title        string `xml:"Title"`
				JournalIssue struct {
					PubDate struct {
						Year  string `xml:"Year"`
						Month string `xml:"Month"`
					} `xml:"PubDate"`
				} `xml:"JournalIssue"`
			} `xml:"Journal"`
			ArticleTitle string `xml:"ArticleTitle"`
			Pagination   struct {
				MedlinePgn string `xml:"MedlinePgn"`
			} `xml:"Pagination"`
			Abstract struct {
				AbstractText string `xml:"AbstractText"`
			} `xml:"Abstract"`
			AuthorList struct {
				Authors []Author `xml:"Author"`
			} `xml:"AuthorList"`
		} `xml:"Article"`
		PMID string `xml:"PMID"`
	} `xml:"MedlineCitation"`
	PubmedData struct {
		ArticleIdList struct {
			ArticleIDs []ArticleID `xml:"ArticleId"`
		} `xml:"ArticleIdList"`
	} `xml:"PubmedData"`
}

// PDFDownloadInfo contains information needed to download a PDF.
type PDFDownloadInfo struct {
	PMID    string
	PMCID   string
	PDFLink *OALink
}
