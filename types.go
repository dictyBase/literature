package literature

import "time"

// Article represents a PubMed article with clean, public-facing fields.
type Article struct {
	PMID        string    `json:"pmid"`
	Title       string    `json:"title"`
	Abstract    string    `json:"abstract"`
	Authors     []Author  `json:"authors"`
	Journal     string    `json:"journal"`
	PublishDate time.Time `json:"publish_date"`
	DOI         string    `json:"doi,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	Volume      string    `json:"volume,omitempty"`
	Issue       string    `json:"issue,omitempty"`
	Pages       string    `json:"pages,omitempty"`
}

// Author represents an article author.
type Author struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
}

// SearchResult represents the result of a literature search.
type SearchResult struct {
	Query    string     `json:"query"`
	Total    int        `json:"total"`
	Articles []*Article `json:"articles"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
}

// PDF represents information about an article's PDF.
type PDF struct {
	PMID     string `json:"pmid"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size,omitempty"`
}

// searchConfig holds internal search configuration.
type searchConfig struct {
	limit  int
	offset int
}

// SearchOption configures search behavior.
type SearchOption func(*searchConfig)

// WithLimit sets the maximum number of search results to return.
func WithLimit(limit int) SearchOption {
	return func(config *searchConfig) {
		if limit > 0 {
			config.limit = limit
		}
	}
}

// WithOffset sets the starting offset for search results.
func WithOffset(offset int) SearchOption {
	return func(config *searchConfig) {
		if offset >= 0 {
			config.offset = offset
		}
	}
}