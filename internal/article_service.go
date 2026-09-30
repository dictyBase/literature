package internal

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ArticleService handles fetching PubMed article metadata.
type ArticleService struct {
	httpClient *http.Client
	baseURL    string
	identity   Identity
}

const eutilsBaseURL = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils"

// NewArticleService creates a new ArticleService with default configuration.
func NewArticleService() *ArticleService {
	return &ArticleService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    eutilsBaseURL,
	}
}

func (s *ArticleService) efetchURL(pmid string) string {
	values := url.Values{}
	values.Set("db", "pubmed")
	values.Set("retmode", "xml")
	values.Set("id", pmid)
	s.identity.Apply(values)

	return s.baseURL + "/efetch.fcgi?" + values.Encode()
}

// FetchArticle retrieves article metadata for the given PMID.
func (s *ArticleService) FetchArticle(pmid string) (*PubMedArticle, error) {
	efetchURL := s.efetchURL(pmid)

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

// formatAuthor formats an author's name as "ForeName LastName".
func formatAuthor(author Author) string {
	return fmt.Sprintf("%s %s", author.ForeName, author.LastName)
}
