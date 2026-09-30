package internal

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SearchService handles PubMed search operations.
type SearchService struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	identity   Identity
}

// SearchServiceOption configures SearchService behavior.
type SearchServiceOption func(*SearchService)

// WithSearchHTTPClient sets a custom HTTP client for the search service.
func WithSearchHTTPClient(client *http.Client) SearchServiceOption {
	return func(s *SearchService) {
		if client != nil {
			s.httpClient = client
		}
	}
}

// WithSearchBaseURL sets the base URL for E-utilities requests.
func WithSearchBaseURL(baseURL string) SearchServiceOption {
	return func(s *SearchService) {
		if baseURL != "" {
			s.baseURL = baseURL
		}
	}
}

// WithSearchUserAgent sets the User-Agent header for E-utilities requests.
func WithSearchUserAgent(userAgent string) SearchServiceOption {
	return func(s *SearchService) {
		if userAgent != "" {
			s.userAgent = userAgent
		}
	}
}

// WithSearchIdentity sets NCBI E-utilities identification parameters.
func WithSearchIdentity(identity Identity) SearchServiceOption {
	return func(service *SearchService) {
		service.identity = identity
	}
}

// NewSearchService creates a new SearchService with the given options.
func NewSearchService(options ...SearchServiceOption) *SearchService {
	service := &SearchService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    eutilsBaseURL,
	}

	for _, option := range options {
		option(service)
	}

	return service
}

func (s *SearchService) esearchRequestURL(
	query string,
	limit, offset int,
) string {
	values := url.Values{}
	values.Set("db", "pubmed")
	values.Set("retmode", "xml")
	values.Set("usehistory", "y")
	values.Set("retmax", strconv.Itoa(limit))
	values.Set("retstart", strconv.Itoa(offset))
	values.Set("term", query)
	s.identity.Apply(values)

	return s.baseURL + "/esearch.fcgi?" + values.Encode()
}

func (s *SearchService) efetchRequestURL(
	webEnv, queryKey string,
	limit, offset int,
) string {
	values := url.Values{}
	values.Set("db", "pubmed")
	values.Set("retmode", "xml")
	values.Set("retmax", strconv.Itoa(limit))
	values.Set("retstart", strconv.Itoa(offset))
	values.Set("WebEnv", webEnv)
	values.Set("query_key", queryKey)
	s.identity.Apply(values)

	return s.baseURL + "/efetch.fcgi?" + values.Encode()
}

// SearchPubMed performs a search query against PubMed and returns search results.
func (s *SearchService) SearchPubMed(
	query string,
	limit, offset int,
) (*ESearchResult, error) {
	esearchURL := s.esearchRequestURL(query, limit, offset)
	request, err := http.NewRequest(http.MethodGet, esearchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating esearch request: %w", redactAPIKey(err))
	}
	if s.userAgent != "" {
		request.Header.Set("User-Agent", s.userAgent)
	}

	// #nosec G107
	resp, err := s.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error making esearch request: %w", redactAPIKey(err))
	}
	defer resp.Body.Close()

	esearchResult := &ESearchResult{}
	if err := xml.NewDecoder(resp.Body).Decode(esearchResult); err != nil {
		return nil, fmt.Errorf("error unmarshaling esearch XML: %w", err)
	}

	return esearchResult, nil
}

// FetchPubMedDetails retrieves detailed article information using WebEnv and QueryKey.
func (s *SearchService) FetchPubMedDetails(
	webEnv, queryKey string,
	limit, offset int,
) (*PubMedArticleSet, error) {
	efetchURL := s.efetchRequestURL(webEnv, queryKey, limit, offset)
	request, err := http.NewRequest(http.MethodGet, efetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating efetch request: %w", redactAPIKey(err))
	}
	if s.userAgent != "" {
		request.Header.Set("User-Agent", s.userAgent)
	}

	// #nosec G107
	resp, err := s.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", redactAPIKey(err))
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling efetch XML: %w", err)
	}

	return articleSet, nil
}
