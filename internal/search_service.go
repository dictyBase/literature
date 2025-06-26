package internal

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SearchService handles PubMed search operations.
type SearchService struct {
	httpClient *http.Client
	baseURL    string
	esearchURL string
	efetchURL  string
	retmax     int
	retstart   int
}

// SearchServiceOption configures SearchService behavior.
type SearchServiceOption func(*SearchService)

// WithSearchHTTPClient sets a custom HTTP client for the search service.
func WithSearchHTTPClient(client *http.Client) SearchServiceOption {
	return func(s *SearchService) {
		s.httpClient = client
	}
}

// WithRetrieval sets the maximum number of results and the starting index.
func WithRetrieval(retmax, retstart int) SearchServiceOption {
	return func(s *SearchService) {
		s.retmax = retmax
		s.retstart = retstart
		s.buildURLs()
	}
}

// NewSearchService creates a new SearchService with the given options.
func NewSearchService(options ...SearchServiceOption) *SearchService {
	service := &SearchService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
		retmax:     10, // default value
		retstart:   0,  // default value
	}

	service.buildURLs()

	for _, option := range options {
		option(service)
	}

	return service
}

// rebuildURLs rebuilds the esearch and efetch URLs with the current retmax value.
func (s *SearchService) buildURLs() {
	s.esearchURL = fmt.Sprintf(
		"%s/esearch.fcgi?db=pubmed&retmax=%d&retstart=%d&retmode=xml&usehistory=y",
		s.baseURL,
		s.retmax,
		s.retstart,
	)
	s.efetchURL = fmt.Sprintf(
		"%s/efetch.fcgi?db=pubmed&retmode=xml&retmax=%d&retstart=%d",
		s.baseURL,
		s.retmax,
		s.retstart,
	)
}

// SearchPubMed performs a search query against PubMed and returns search results.
func (s *SearchService) SearchPubMed(
	query string,
	limit, offset int,
) (*ESearchResult, error) {
	s.retmax = limit
	s.retstart = offset
	s.buildURLs()
	esearchURL := fmt.Sprintf(
		"%s&term=%s",
		s.esearchURL,
		url.QueryEscape(query),
	)

	// #nosec G107
	resp, err := s.httpClient.Get(esearchURL)
	if err != nil {
		return nil, fmt.Errorf("error making esearch request: %w", err)
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
) (*PubMedArticleSet, error) {
	efetchURL := fmt.Sprintf(
		"%s&WebEnv=%s&query_key=%s",
		s.efetchURL,
		webEnv,
		queryKey,
	)

	// #nosec G107
	resp, err := s.httpClient.Get(efetchURL)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", err)
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling efetch XML: %w", err)
	}

	return articleSet, nil
}
