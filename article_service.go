package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// ArticleService handles fetching PubMed article metadata.
type ArticleService struct {
	httpClient *http.Client
	baseURL    string
}

// NewArticleService creates a new ArticleService with default configuration.
func NewArticleService() *ArticleService {
	return &ArticleService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
	}
}

// FetchArticle retrieves article metadata for the given PMID.
func (s *ArticleService) FetchArticle(pmid string) (*PubMedArticle, error) {
	efetchURL := fmt.Sprintf(
		"%s/efetch.fcgi?db=pubmed&retmode=xml&id=%s",
		s.baseURL,
		pmid,
	)

	// #nosec G107
	resp, err := s.httpClient.Get(efetchURL)
	if err != nil {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorArticleNotFound,
			Err:  fmt.Errorf("error making efetch request: %w", err),
		}
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorArticleNotFound,
			Err:  fmt.Errorf("error unmarshaling efetch XML: %w", err),
		}
	}

	if len(articleSet.PubMedArticles) == 0 {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorArticleNotFound,
			Err:  fmt.Errorf("no articles found"),
		}
	}

	return &articleSet.PubMedArticles[0], nil
}
