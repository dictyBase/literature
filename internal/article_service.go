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
	userAgent  string
	identity   Identity
}

// ArticleServiceOption configures ArticleService behavior.
type ArticleServiceOption func(*ArticleService)

const eutilsBaseURL = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils"

// WithArticleHTTPClient sets a custom HTTP client for the article service.
func WithArticleHTTPClient(client *http.Client) ArticleServiceOption {
	return func(service *ArticleService) {
		if client != nil {
			service.httpClient = client
		}
	}
}

// WithArticleBaseURL sets the base URL for E-utilities requests.
func WithArticleBaseURL(baseURL string) ArticleServiceOption {
	return func(service *ArticleService) {
		if baseURL != "" {
			service.baseURL = baseURL
		}
	}
}

// WithArticleUserAgent sets the User-Agent header for E-utilities requests.
func WithArticleUserAgent(userAgent string) ArticleServiceOption {
	return func(service *ArticleService) {
		if userAgent != "" {
			service.userAgent = userAgent
		}
	}
}

// WithArticleIdentity sets NCBI E-utilities identification parameters.
func WithArticleIdentity(identity Identity) ArticleServiceOption {
	return func(service *ArticleService) {
		service.identity = identity
	}
}

// NewArticleService creates a new ArticleService with default configuration.
func NewArticleService(options ...ArticleServiceOption) *ArticleService {
	service := &ArticleService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    eutilsBaseURL,
	}

	for _, option := range options {
		option(service)
	}

	return service
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
	request, err := http.NewRequest(http.MethodGet, efetchURL, nil)
	if err != nil {
		return nil, &PDFError{
			PMID: pmid,
			Type: PDFErrorArticleNotFound,
			Err:  fmt.Errorf("error creating efetch request: %w", err),
		}
	}
	if s.userAgent != "" {
		request.Header.Set("User-Agent", s.userAgent)
	}

	// #nosec G107
	resp, err := s.httpClient.Do(request)
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
