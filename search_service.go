package main

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
}

// SearchServiceOption configures SearchService behavior.
type SearchServiceOption func(*SearchService)

// WithSearchHTTPClient sets a custom HTTP client for the search service.
func WithSearchHTTPClient(client *http.Client) SearchServiceOption {
	return func(s *SearchService) {
		s.httpClient = client
	}
}

// NewSearchService creates a new SearchService with the given options.
func NewSearchService(options ...SearchServiceOption) *SearchService {
	service := &SearchService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
	}

	for _, option := range options {
		option(service)
	}

	return service
}

// SearchPubMed performs a search query against PubMed and returns search results.
func (s *SearchService) SearchPubMed(query string) (*ESearchResult, error) {
	esearchURL := fmt.Sprintf(
		"%s/esearch.fcgi?db=pubmed&term=%s&retmax=10&retmode=xml&usehistory=y",
		s.baseURL,
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
func (s *SearchService) FetchPubMedDetails(webEnv, queryKey string) (*PubMedArticleSet, error) {
	efetchURL := fmt.Sprintf(
		"%s/efetch.fcgi?db=pubmed&retmode=xml&WebEnv=%s&query_key=%s&retmax=10",
		s.baseURL,
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
