package literature

import "time"

// EuropePMCArticle represents an article from EuropePMC with clean, public-facing fields.
type EuropePMCArticle struct {
	ID           string                 `json:"id"`
	Source       string                 `json:"source"`
	PMID         string                 `json:"pmid"`
	PMCID        string                 `json:"pmcid,omitempty"`
	DOI          string                 `json:"doi,omitempty"`
	Title        string                 `json:"title"`
	AuthorString string                 `json:"author_string"`
	Authors      []EuropePMCAuthor      `json:"authors"`
	Abstract     string                 `json:"abstract"`
	Journal      EuropePMCJournal       `json:"journal"`
	PubYear      string                 `json:"pub_year"`
	PageInfo     string                 `json:"page_info,omitempty"`
	Keywords     []string               `json:"keywords,omitempty"`
	FullTextURLs []EuropePMCFullTextURL `json:"full_text_urls,omitempty"`
	IsOpenAccess bool                   `json:"is_open_access"`
	HasPDF       bool                   `json:"has_pdf"`
	License      string                 `json:"license,omitempty"`
	CitedByCount int                    `json:"cited_by_count"`
	Language     string                 `json:"language,omitempty"`
	PubTypes     []string               `json:"pub_types,omitempty"`
	MeshHeadings []EuropePMCMeshHeading `json:"mesh_headings,omitempty"`
	Chemicals    []EuropePMCChemical    `json:"chemicals,omitempty"`
	Grants       []EuropePMCGrant       `json:"grants,omitempty"`
	PublishDate  *time.Time             `json:"publish_date,omitempty"`
	CreationDate *time.Time             `json:"creation_date,omitempty"`
	RevisionDate *time.Time             `json:"revision_date,omitempty"`
}

// EuropePMCAuthor represents an article author with detailed information.
type EuropePMCAuthor struct {
	FullName     string                       `json:"full_name"`
	FirstName    string                       `json:"first_name"`
	LastName     string                       `json:"last_name"`
	Initials     string                       `json:"initials"`
	ORCID        string                       `json:"orcid,omitempty"`
	Affiliations []EuropePMCAuthorAffiliation `json:"affiliations,omitempty"`
}

// EuropePMCAuthorAffiliation represents an author's institutional affiliation.
type EuropePMCAuthorAffiliation struct {
	Affiliation string `json:"affiliation"`
}

// EuropePMCJournal represents journal information with detailed metadata.
type EuropePMCJournal struct {
	Title               string `json:"title"`
	MedlineAbbreviation string `json:"medline_abbreviation,omitempty"`
	ISOAbbreviation     string `json:"iso_abbreviation,omitempty"`
	ISSN                string `json:"issn,omitempty"`
	ESSN                string `json:"essn,omitempty"`
	Volume              string `json:"volume,omitempty"`
	Issue               string `json:"issue,omitempty"`
	IssueID             int    `json:"issue_id,omitempty"`
	DateOfPublication   string `json:"date_of_publication,omitempty"`
	MonthOfPublication  int    `json:"month_of_publication,omitempty"`
	YearOfPublication   int    `json:"year_of_publication,omitempty"`
	NLMID               string `json:"nlm_id,omitempty"`
}

// EuropePMCFullTextURL represents a URL to access the full text of an article.
type EuropePMCFullTextURL struct {
	Availability     string `json:"availability"`
	AvailabilityCode string `json:"availability_code"`
	DocumentStyle    string `json:"document_style"`
	Site             string `json:"site"`
	URL              string `json:"url"`
}

// EuropePMCMeshHeading represents a MeSH (Medical Subject Heading) term.
type EuropePMCMeshHeading struct {
	MajorTopic     bool                     `json:"major_topic"`
	DescriptorName string                   `json:"descriptor_name"`
	MeshQualifiers []EuropePMCMeshQualifier `json:"mesh_qualifiers,omitempty"`
}

// EuropePMCMeshQualifier represents a MeSH qualifier.
type EuropePMCMeshQualifier struct {
	Abbreviation  string `json:"abbreviation"`
	QualifierName string `json:"qualifier_name"`
	MajorTopic    bool   `json:"major_topic"`
}

// EuropePMCChemical represents a chemical substance mentioned in the article.
type EuropePMCChemical struct {
	Name           string `json:"name"`
	RegistryNumber string `json:"registry_number,omitempty"`
}

// EuropePMCGrant represents funding information for the research.
type EuropePMCGrant struct {
	GrantID string `json:"grant_id"`
	Agency  string `json:"agency"`
	OrderIn int    `json:"order_in"`
}

// EuropePMCSearchResult represents the result of a literature search in EuropePMC.
type EuropePMCSearchResult struct {
	Query    string              `json:"query"`
	Total    int                 `json:"total"`
	Articles []*EuropePMCArticle `json:"articles"`
	Limit    int                 `json:"limit"`
	Offset   int                 `json:"offset"`
}

// europePMCSearchConfig holds internal search configuration.
type europePMCSearchConfig struct {
	limit      int
	offset     int
	resultType string
	format     string
}

// EuropePMCSearchOption configures search behavior.
type EuropePMCSearchOption func(*europePMCSearchConfig)

// WithEuropePMCLimit sets the maximum number of search results to return.
func WithEuropePMCLimit(limit int) EuropePMCSearchOption {
	return func(config *europePMCSearchConfig) {
		if limit > 0 {
			config.limit = limit
		}
	}
}

// WithEuropePMCOffset sets the starting offset for search results.
func WithEuropePMCOffset(offset int) EuropePMCSearchOption {
	return func(config *europePMCSearchConfig) {
		if offset >= 0 {
			config.offset = offset
		}
	}
}

// WithEuropePMCResultType sets the result type (core, lite, etc.).
func WithEuropePMCResultType(resultType string) EuropePMCSearchOption {
	return func(config *europePMCSearchConfig) {
		if resultType != "" {
			config.resultType = resultType
		}
	}
}

// WithEuropePMCFormat sets the response format (json, xml).
func WithEuropePMCFormat(format string) EuropePMCSearchOption {
	return func(config *europePMCSearchConfig) {
		if format != "" {
			config.format = format
		}
	}
}
