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

// GetIDs returns the list of PubMed IDs from the search result.
func (e *ESearchResult) GetIDs() []string {
	return e.IDList.IDs
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

// PDFDownloadInfo contains information needed to download a PDF.
type PDFDownloadInfo struct {
	PMID    string
	PMCID   string
	PDFLink *OALink
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

// Convenience methods for PubMedArticle

// GetPMID returns the PubMed ID.
func (p *PubMedArticle) GetPMID() string {
	return p.MedlineCitation.PMID
}

// GetTitle returns the article title.
func (p *PubMedArticle) GetTitle() string {
	return p.MedlineCitation.Article.ArticleTitle
}

// GetJournalTitle returns the journal title.
func (p *PubMedArticle) GetJournalTitle() string {
	return p.MedlineCitation.Article.Journal.Title
}

// GetAbstract returns the article abstract.
func (p *PubMedArticle) GetAbstract() string {
	return p.MedlineCitation.Article.Abstract.AbstractText
}

// GetAuthors returns the list of authors.
func (p *PubMedArticle) GetAuthors() []Author {
	return p.MedlineCitation.Article.AuthorList.Authors
}

// GetPages returns the page range.
func (p *PubMedArticle) GetPages() string {
	return p.MedlineCitation.Article.Pagination.MedlinePgn
}

// GetPubYear returns the publication year.
func (p *PubMedArticle) GetPubYear() string {
	return p.MedlineCitation.Article.Journal.JournalIssue.PubDate.Year
}

// GetPubMonth returns the publication month.
func (p *PubMedArticle) GetPubMonth() string {
	return p.MedlineCitation.Article.Journal.JournalIssue.PubDate.Month
}

// GetArticleIDs returns the list of article identifiers.
func (p *PubMedArticle) GetArticleIDs() []ArticleID {
	return p.PubmedData.ArticleIdList.ArticleIDs
}

// GetDOI returns the DOI if available.
func (p *PubMedArticle) GetDOI() (string, bool) {
	doiArticleID, found := Find(p.GetArticleIDs(), isDOI)
	if found {
		return doiArticleID.Value, true
	}
	return "", false
}

// GetPMCID returns the PMC ID if available.
func (p *PubMedArticle) GetPMCID() (string, bool) {
	pmcArticleID, found := Find(p.GetArticleIDs(), isPMCID)
	if found {
		return pmcArticleID.Value, true
	}
	return "", false
}
