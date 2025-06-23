package main

import (
	"encoding/xml"
)

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
	doiArticleID, found := Find(p.GetArticleIDs(), IsDOI)
	if found {
		return doiArticleID.Value, true
	}
	return "", false
}

// GetPMCID returns the PMC ID if available.
func (p *PubMedArticle) GetPMCID() (string, bool) {
	pmcArticleID, found := Find(p.GetArticleIDs(), IsPMCID)
	if found {
		return pmcArticleID.Value, true
	}
	return "", false
}
