package literature

import (
	"fmt"
	"net/http"
	"strings"
)

// Internal service adapters that use the existing internal implementations

// articleService wraps the internal article service
type articleService struct {
	httpClient *http.Client
	baseURL    string
}

// newArticleService creates a new articleService
func newArticleService(httpClient *http.Client, baseURL string) *articleService {
	return &articleService{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// fetchArticle is a placeholder - would use the internal service
func (s *articleService) fetchArticle(pmid string) (*internalArticle, error) {
	// This is a simplified placeholder
	return &internalArticle{
		PMID:     pmid,
		Title:    "Sample Article Title",
		Abstract: "Sample abstract",
		Authors:  []internalAuthor{{FirstName: "John", LastName: "Doe"}},
		Journal:  "Sample Journal",
	}, nil
}

// searchService wraps the internal search service
type searchService struct {
	httpClient *http.Client
	baseURL    string
}

// newSearchService creates a new searchService
func newSearchService(httpClient *http.Client, baseURL string) *searchService {
	return &searchService{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// search is a placeholder - would use the internal service
func (s *searchService) search(query string, limit, offset int) (*internalSearchResult, error) {
	// This is a simplified placeholder
	return &internalSearchResult{
		Query:  query,
		Total:  100,
		PMIDs:  []string{"12345678", "87654321"},
		Limit:  limit,
		Offset: offset,
	}, nil
}

// pdfService wraps the internal PDF service
type pdfService struct {
	httpClient *http.Client
	baseURL    string
}

// newPDFService creates a new pdfService
func newPDFService(httpClient *http.Client, baseURL string) *pdfService {
	return &pdfService{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// getPDF is a placeholder - would use the internal service
func (s *pdfService) getPDF(pmid string) (*internalPDF, error) {
	// This is a simplified placeholder
	return &internalPDF{
		PMID:     pmid,
		URL:      fmt.Sprintf("https://example.com/pdf/%s.pdf", pmid),
		Filename: fmt.Sprintf("article_%s.pdf", pmid),
	}, nil
}

// hasPDF is a placeholder - would use the internal service
func (s *pdfService) hasPDF(pmid string) (bool, error) {
	// This is a simplified placeholder
	return true, nil
}

// Internal data structures (simplified)

type internalArticle struct {
	PMID     string
	Title    string
	Abstract string
	Authors  []internalAuthor
	Journal  string
	DOI      string
}

type internalAuthor struct {
	FirstName string
	LastName  string
}

type internalSearchResult struct {
	Query  string
	Total  int
	PMIDs  []string
	Limit  int
	Offset int
}

type internalPDF struct {
	PMID     string
	URL      string
	Filename string
}

// Conversion functions

func convertToPublicArticle(internal *internalArticle) *Article {
	article := &Article{
		PMID:     internal.PMID,
		Title:    internal.Title,
		Abstract: internal.Abstract,
		Journal:  internal.Journal,
		DOI:      internal.DOI,
	}

	// Convert authors
	for _, author := range internal.Authors {
		article.Authors = append(article.Authors, Author{
			FirstName: author.FirstName,
			LastName:  author.LastName,
			FullName:  strings.TrimSpace(fmt.Sprintf("%s %s", author.FirstName, author.LastName)),
		})
	}

	return article
}

func convertToPublicSearchResult(internal *internalSearchResult, query string) *SearchResult {
	// For now, we'll just return the structure without full articles
	// In a complete implementation, you'd fetch each PMID
	return &SearchResult{
		Query:    query,
		Total:    internal.Total,
		Limit:    internal.Limit,
		Offset:   internal.Offset,
		Articles: []*Article{}, // Would be populated by fetching PMIDs
	}
}

func convertToPublicPDF(internal *internalPDF) *PDF {
	return &PDF{
		PMID:     internal.PMID,
		URL:      internal.URL,
		Filename: internal.Filename,
	}
}