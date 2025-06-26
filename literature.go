// Package literature provides a Go client for accessing PubMed literature data.
// It offers a clean API for searching articles, fetching metadata, and accessing PDFs.
package literature

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dictybase/literature/internal"
)

// Client provides access to PubMed literature services.
type Client struct {
	articleService *internal.ArticleService
	searchService  *internal.SearchService
	pdfService     *internal.PDFService
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

	// Initialize services using internal constructors with appropriate options
	var pdfOpts []internal.PDFServiceOption
	var searchOpts []internal.SearchServiceOption

	// Pass HTTP client to internal services if configured
	if client.httpClient != nil {
		pdfOpts = append(pdfOpts, internal.WithHTTPClient(client.httpClient))
		searchOpts = append(
			searchOpts,
			internal.WithSearchHTTPClient(client.httpClient),
		)
	}

	client.articleService = internal.NewArticleService()
	client.searchService = internal.NewSearchService(searchOpts...)
	client.pdfService = internal.NewPDFService(pdfOpts...)

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

	internalArticle, err := c.articleService.FetchArticle(pmid)
	if err != nil {
		return nil, err
	}

	return convertFromInternalArticle(internalArticle), nil
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
func (c *Client) Search(
	query string,
	opts ...SearchOption,
) (*SearchResult, error) {
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

	searchResult, err := c.searchService.SearchPubMed(query)
	if err != nil {
		return nil, err
	}

	// Fetch detailed article information using WebEnv and QueryKey
	articleSet, err := c.searchService.FetchPubMedDetails(
		searchResult.WebEnv,
		searchResult.QueryKey,
	)
	if err != nil {
		return nil, err
	}

	return convertFromInternalSearchResultWithArticles(
		searchResult,
		articleSet,
		query,
		config.limit,
		config.offset,
	), nil
}

// FindSimilar finds articles similar to the given PMID.
func (c *Client) FindSimilar(
	pmid string,
	opts ...SearchOption,
) (*SearchResult, error) {
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

	available, err := c.pdfService.IsPDFAvailable(pmid)
	if err != nil {
		return nil, err
	}

	if !available {
		return nil, &Error{
			Type:    ErrorTypePDFNotAvailable,
			Message: "PDF not available for this article",
		}
	}

	url, err := c.pdfService.GetPDFURL()
	if err != nil {
		return nil, err
	}

	return &PDF{
		PMID: pmid,
		URL:  url,
	}, nil
}

// HasPDF checks if a PDF is available for the given PMID.
func (c *Client) HasPDF(pmid string) (bool, error) {
	if pmid == "" {
		return false, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: "PMID cannot be empty",
		}
	}

	return c.pdfService.IsPDFAvailable(pmid)
}
