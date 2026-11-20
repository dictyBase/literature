package literature

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dictybase/literature/internal"
	"github.com/go-playground/validator/v10"
)

// Client provides access to PubMed literature services.
type Client struct {
	articleService *internal.ArticleService
	searchService  *internal.SearchService
	pdfService     *internal.PDFService
	httpClient     *http.Client
	baseURL        string
	userAgent      string
	apiKey         string
	tool           string
	email          string
	rateLimit      float64
	maxRetries     int
	retryDelay     time.Duration
	validate       *validator.Validate
}

// NCBI request limits for E-utilities: 3 requests per second without an API
// key and 10 with one. A 429 response gets defaultMaxRetries retries.
const (
	rateLimitWithoutKey = 3
	rateLimitWithKey    = 10
	defaultMaxRetries   = 3
	defaultRetryDelay   = time.Second
)

// New creates a new literature client with the provided options.
func New(opts ...Option) (*Client, error) {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
		userAgent:  "literature-go-client/1.0",
		maxRetries: defaultMaxRetries,
		retryDelay: defaultRetryDelay,
		validate:   validator.New(),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	identity := internal.Identity{
		APIKey: client.apiKey,
		Tool:   client.tool,
		Email:  client.email,
	}

	ncbiClient := client.ncbiHTTPClient()
	client.articleService = internal.NewArticleService(
		internal.WithArticleHTTPClient(ncbiClient),
		internal.WithArticleIdentity(identity),
		internal.WithArticleBaseURL(client.baseURL),
		internal.WithArticleUserAgent(client.userAgent),
	)
	client.searchService = internal.NewSearchService(
		internal.WithSearchHTTPClient(ncbiClient),
		internal.WithSearchIdentity(identity),
		internal.WithSearchBaseURL(client.baseURL),
		internal.WithSearchUserAgent(client.userAgent),
	)
	client.pdfService = internal.NewPDFService(
		internal.WithPDFHTTPClient(client.httpClient),
		internal.WithPDFArticleService(client.articleService),
	)

	return client, nil
}

// ncbiHTTPClient copies the HTTP client and wraps its transport with the
// shared NCBI rate limiter and the 429 retry. Only the E-utilities services
// use it. PDF downloads go to other hosts and keep the plain client.
func (c *Client) ncbiHTTPClient() *http.Client {
	requestsPerSecond := c.rateLimit
	if requestsPerSecond <= 0 {
		requestsPerSecond = rateLimitWithoutKey
		if c.apiKey != "" {
			requestsPerSecond = rateLimitWithKey
		}
	}
	ncbi := *c.httpClient
	ncbi.Transport = &internal.NCBITransport{
		Base:       c.httpClient.Transport,
		Limiter:    internal.SharedLimiter(c.baseURL, c.apiKey, requestsPerSecond),
		MaxRetries: c.maxRetries,
		RetryDelay: c.retryDelay,
	}
	return &ncbi
}

// GetArticle retrieves article metadata for the given PMID.
func (c *Client) GetArticle(pmid string) (*Article, error) {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	internalArticle, err := c.articleService.FetchArticle(pmid)
	if err != nil {
		return nil, err
	}

	return convertFromInternalArticle(internalArticle), nil
}

// fetchArticleWithError is a helper function that wraps GetArticle for
// functional composition.
func (c *Client) fetchArticleWithError(pmid string) (*Article, error) {
	article, err := c.GetArticle(pmid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article %s: %w", pmid, err)
	}
	return article, nil
}

// GetArticles retrieves metadata for multiple PMIDs.
func (c *Client) GetArticles(pmids []string) ([]*Article, error) {
	if err := c.validate.Var(pmids, "required,min=1,dive,required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	return internal.MapWithError(pmids, c.fetchArticleWithError)
}

// Search performs a literature search with the given query.
func (c *Client) Search(
	query string,
	opts ...SearchOption,
) (*SearchResult, error) {
	if err := c.validate.Var(query, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}
	config := &searchConfig{
		limit:  20,
		offset: 0,
	}

	for _, opt := range opts {
		opt(config)
	}

	searchResult, err := c.searchService.SearchPubMed(
		query,
		config.limit,
		config.offset,
	)
	if err != nil {
		return nil, err
	}

	// Fetch detailed article information using WebEnv and QueryKey
	articleSet, err := c.searchService.FetchPubMedDetails(
		searchResult.WebEnv,
		searchResult.QueryKey,
		config.limit,
		config.offset,
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
	if err := c.validate.Var(pmid, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
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
	query := fmt.Sprintf("pubmed_similar[PMID]:%s", pmid)
	return c.Search(query, WithLimit(config.limit), WithOffset(config.offset))
}

// GetPDF retrieves PDF information for the given PMID.
func (c *Client) GetPDF(pmid string) (*PDF, error) {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
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
	if err := c.validate.Var(pmid, "required"); err != nil {
		return false, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	return c.pdfService.IsPDFAvailable(pmid)
}

// DownloadPDF downloads the full PDF for the given PMID to the specified file path.
func (c *Client) DownloadPDF(pmid, filePath string) error {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed for pmid: %s", err.Error()),
		}
	}
	if err := c.validate.Var(filePath, "required"); err != nil {
		return &Error{
			Type: ErrorTypeInvalidInput,
			Message: fmt.Sprintf(
				"validation failed for filePath: %s",
				err.Error(),
			),
		}
	}

	return c.pdfService.DownloadArticlePDF(pmid, filePath)
}
