// Package literature provides a Go client for accessing PubMed literature data.
// It offers a clean API for searching articles, fetching metadata, and accessing PDFs.
package literature

import (
	"fmt"
	"net/http"
	"time"
)

// Client provides access to PubMed literature services.
type Client struct {
	articleService *articleService
	searchService  *searchService
	pdfService     *pdfService
	httpClient     *http.Client
	baseURL        string
	userAgent      string
}

// New creates a new literature client with the provided options.
func New(opts ...Option) (*Client, error) {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
		userAgent:  "literature-go-client/1.0",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Initialize services
	client.articleService = newArticleService(client.httpClient, client.baseURL)
	client.searchService = newSearchService(client.httpClient, client.baseURL)
	client.pdfService = newPDFService(client.httpClient, client.baseURL)

	return client, nil
}

// GetArticle retrieves article metadata for the given PMID.
func (c *Client) GetArticle(pmid string) (*Article, error) {
	if pmid == "" {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "PMID cannot be empty",
		}
	}

	internalArticle, err := c.articleService.fetchArticle(pmid)
	if err != nil {
		return nil, err
	}

	return convertToPublicArticle(internalArticle), nil
}

// GetArticles retrieves metadata for multiple PMIDs.
func (c *Client) GetArticles(pmids []string) ([]*Article, error) {
	if len(pmids) == 0 {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "PMIDs list cannot be empty",
		}
	}

	articles := make([]*Article, 0, len(pmids))
	for _, pmid := range pmids {
		article, err := c.GetArticle(pmid)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch article %s: %w", pmid, err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// Search performs a literature search with the given query.
func (c *Client) Search(query string, opts ...SearchOption) (*SearchResult, error) {
	if query == "" {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "search query cannot be empty",
		}
	}

	config := &searchConfig{
		limit:  20,
		offset: 0,
	}

	for _, opt := range opts {
		opt(config)
	}

	internalResult, err := c.searchService.search(query, config.limit, config.offset)
	if err != nil {
		return nil, err
	}

	return convertToPublicSearchResult(internalResult, query), nil
}

// FindSimilar finds articles similar to the given PMID.
func (c *Client) FindSimilar(pmid string, opts ...SearchOption) (*SearchResult, error) {
	if pmid == "" {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "PMID cannot be empty",
		}
	}

	config := &searchConfig{
		limit:  20,
		offset: 0,
	}

	for _, opt := range opts {
		opt(config)
	}

	// Use PubMed's "similar articles" query
	query := fmt.Sprintf("similar_articles[PMID]:%s", pmid)
	return c.Search(query, WithLimit(config.limit), WithOffset(config.offset))
}

// GetPDF retrieves PDF information for the given PMID.
func (c *Client) GetPDF(pmid string) (*PDF, error) {
	if pmid == "" {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "PMID cannot be empty",
		}
	}

	internalPDF, err := c.pdfService.getPDF(pmid)
	if err != nil {
		return nil, err
	}

	return convertToPublicPDF(internalPDF), nil
}

// HasPDF checks if a PDF is available for the given PMID.
func (c *Client) HasPDF(pmid string) (bool, error) {
	if pmid == "" {
		return false, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "PMID cannot be empty",
		}
	}

	return c.pdfService.hasPDF(pmid)
}