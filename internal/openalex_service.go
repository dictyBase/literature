package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// openAlexBaseURL is the default OpenAlex API base URL.
	openAlexBaseURL = "https://api.openalex.org"

	// openAlexDOIURLPrefix is the canonical DOI URL prefix used by OpenAlex.
	openAlexDOIURLPrefix = "https://doi.org/"
)

// OpenAlexService handles API interactions with OpenAlex.
type OpenAlexService struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	email      string
	apiKey     string
}

// OpenAlexServiceOption configures the OpenAlex service.
type OpenAlexServiceOption func(*OpenAlexService)

// WithOpenAlexHTTPClient sets the HTTP client for the service.
func WithOpenAlexHTTPClient(client *http.Client) OpenAlexServiceOption {
	return func(s *OpenAlexService) {
		if client != nil {
			s.httpClient = client
		}
	}
}

// WithOpenAlexBaseURL sets the base URL for the service (primarily for testing).
func WithOpenAlexBaseURL(baseURL string) OpenAlexServiceOption {
	return func(s *OpenAlexService) {
		if baseURL != "" {
			s.baseURL = baseURL
		}
	}
}

// WithOpenAlexUserAgent sets the User-Agent header.
func WithOpenAlexUserAgent(userAgent string) OpenAlexServiceOption {
	return func(s *OpenAlexService) {
		if userAgent != "" {
			s.userAgent = userAgent
		}
	}
}

// WithOpenAlexEmail sets the email for the OpenAlex "Polite Pool".
// The email is sent as the mailto query parameter on every request.
func WithOpenAlexEmail(email string) OpenAlexServiceOption {
	return func(s *OpenAlexService) {
		if email != "" {
			s.email = email
		}
	}
}

// WithOpenAlexAPIKey sets the OpenAlex Premium API key.
// The key is sent as the api_key query parameter on every request.
func WithOpenAlexAPIKey(apiKey string) OpenAlexServiceOption {
	return func(s *OpenAlexService) {
		if apiKey != "" {
			s.apiKey = apiKey
		}
	}
}

// NewOpenAlexService creates a new OpenAlex service with the provided options.
func NewOpenAlexService(opts ...OpenAlexServiceOption) *OpenAlexService {
	service := &OpenAlexService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    openAlexBaseURL,
		userAgent:  "openalex-go-client/1.0",
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

// OpenAlexWorksParams holds parameters for OpenAlex works list requests.
type OpenAlexWorksParams struct {
	Filter  string
	PerPage int
	Cursor  string
	Sort    string
}

// FetchWork retrieves a single work by its OpenAlex ID (W...), PMID
// identifier (PMID:12345678), or DOI URL.
func (s *OpenAlexService) FetchWork(id string) (*OpenAlexAPIWork, error) {
	workURL, err := s.buildWorkURL(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build work URL: %w", err)
	}

	return s.getWork(workURL)
}

// SearchWorks performs a works list request against the OpenAlex API.
func (s *OpenAlexService) SearchWorks(
	params OpenAlexWorksParams,
) (*OpenAlexAPIWorksResponse, error) {
	worksURL, err := s.buildWorksURL(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build works URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, worksURL, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create request: %w",
			redactAPIKey(err),
		)
	}
	s.setCommonHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, s.worksListError(resp.StatusCode)
	}

	apiResponse := &OpenAlexAPIWorksResponse{}
	if err := json.NewDecoder(resp.Body).Decode(apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResponse, nil
}

// buildWorkURL constructs the single-work URL for the given identifier.
// Identifiers may be an OpenAlex ID (W...), a full OpenAlex URL
// (https://openalex.org/W...), a PMID (PMID:12345678), a plain DOI, or a
// full DOI URL.
func (s *OpenAlexService) buildWorkURL(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", fmt.Errorf("empty work identifier")
	}

	pathSegment := stripOpenAlexURLPrefix(trimmed)
	if isDOIIdentifier(trimmed) {
		pathSegment = openAlexDOIURLPrefix + trimDOIPrefix(trimmed)
	}

	base, err := url.Parse(s.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	parsed := *base
	parsed.Path = strings.TrimSuffix(base.Path, "/") +
		"/works/" + pathSegment

	query := parsed.Query()
	s.applyCredentials(query)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

// stripOpenAlexURLPrefix trims the https://openalex.org/ prefix from a
// full OpenAlex entity URL, e.g. https://openalex.org/W123 -> W123.
func stripOpenAlexURLPrefix(identifier string) string {
	const prefix = "https://openalex.org/"
	if strings.HasPrefix(strings.ToLower(identifier), prefix) {
		return identifier[len(prefix):]
	}
	return identifier
}

// buildWorksURL constructs the works list URL with query parameters.
func (s *OpenAlexService) buildWorksURL(
	params OpenAlexWorksParams,
) (string, error) {
	base, err := url.Parse(s.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	parsed := *base
	parsed.Path = strings.TrimSuffix(base.Path, "/") + "/works"

	query := parsed.Query()
	if params.Filter != "" {
		query.Set("filter", params.Filter)
	}
	if params.PerPage > 0 {
		query.Set("per-page", strconv.Itoa(params.PerPage))
	}
	if params.Cursor != "" {
		query.Set("cursor", params.Cursor)
	}
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	}
	s.applyCredentials(query)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

// applyCredentials adds the polite pool mailto and premium api_key
// parameters when configured.
func (s *OpenAlexService) applyCredentials(query url.Values) {
	if s.email != "" {
		query.Set("mailto", s.email)
	}
	if s.apiKey != "" {
		query.Set("api_key", s.apiKey)
	}
}

// setCommonHeaders sets the shared HTTP headers for OpenAlex requests.
func (s *OpenAlexService) setCommonHeaders(req *http.Request) {
	if s.userAgent != "" {
		req.Header.Set("User-Agent", s.userAgent)
	}
	req.Header.Set("Accept", "application/json")
}

// getWork performs a GET request expecting a single work JSON object.
func (s *OpenAlexService) getWork(workURL string) (*OpenAlexAPIWork, error) {
	req, err := http.NewRequest(http.MethodGet, workURL, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create request: %w",
			redactAPIKey(err),
		)
	}
	s.setCommonHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, s.worksListError(resp.StatusCode)
	}

	work := &OpenAlexAPIWork{}
	if err := json.NewDecoder(resp.Body).Decode(work); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return work, nil
}

// Sentinel errors returned by the OpenAlex service for well-known API
// status codes. Use errors.Is to map them to public client errors.
var (
	// ErrOpenAlexNotFound indicates the requested work does not exist.
	ErrOpenAlexNotFound = errors.New("work not found")

	// ErrOpenAlexRateLimited indicates the API rate limit was exceeded.
	ErrOpenAlexRateLimited = errors.New("rate limit exceeded")
)

// worksListError maps a non-200 status code to an error message.
func (s *OpenAlexService) worksListError(statusCode int) error {
	if statusCode == http.StatusNotFound {
		return fmt.Errorf("%w (status %d)", ErrOpenAlexNotFound, statusCode)
	}
	if statusCode == http.StatusTooManyRequests {
		return fmt.Errorf("%w (status %d)", ErrOpenAlexRateLimited, statusCode)
	}
	return fmt.Errorf("API request failed with status %d", statusCode)
}

// isDOIIdentifier reports whether the identifier refers to a DOI rather
// than an OpenAlex ID or PMID.
func isDOIIdentifier(identifier string) bool {
	lower := strings.ToLower(identifier)
	if strings.HasPrefix(lower, openAlexDOIURLPrefix) {
		return true
	}
	if strings.HasPrefix(lower, "doi:") {
		return true
	}
	return strings.HasPrefix(identifier, "10.") &&
		strings.Contains(identifier, "/")
}

// trimDOIPrefix normalizes the identifier to a bare DOI path.
func trimDOIPrefix(id string) string {
	trimmed := strings.TrimSpace(id)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, openAlexDOIURLPrefix) {
		return trimmed[len(openAlexDOIURLPrefix):]
	}
	if strings.HasPrefix(lower, "doi:") {
		return strings.TrimPrefix(trimmed[4:], "//")
	}
	return trimmed
}
