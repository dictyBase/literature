package literature

// OpenAlexWork represents a scholarly work from OpenAlex with clean,
// public-facing fields.
type OpenAlexWork struct {
	OpenAlexID      string               `json:"openalex_id"`
	DOI             string               `json:"doi,omitempty"`
	PMID            string               `json:"pmid,omitempty"`
	Title           string               `json:"title"`
	PublicationYear int                  `json:"publication_year"`
	PublicationDate string               `json:"publication_date,omitempty"`
	Type            string               `json:"type,omitempty"`
	CitedByCount    int                  `json:"cited_by_count"`
	FWCI            float64              `json:"fwci"`
	PercentileMin   float64              `json:"percentile_min"`
	PercentileMax   float64              `json:"percentile_max"`
	ReferencedWorks []string             `json:"referenced_works,omitempty"`
	RelatedWorks    []string             `json:"related_works,omitempty"`
	IsRetracted     bool                 `json:"is_retracted"`
	OpenAccess      OpenAlexOpenAccess   `json:"open_access"`
	PrimaryLocation OpenAlexLocation     `json:"primary_location"`
	BestOALocation  *OpenAlexLocation    `json:"best_oa_location,omitempty"`
	Authors         []OpenAlexAuthorship `json:"authors"`
	Topics          []OpenAlexTopic      `json:"topics"`
	Keywords        []OpenAlexKeyword    `json:"keywords,omitempty"`
	Awards          []OpenAlexAward      `json:"awards,omitempty"`
	Volume          string               `json:"volume,omitempty"`
	Issue           string               `json:"issue,omitempty"`
	FirstPage       string               `json:"first_page,omitempty"`
	LastPage        string               `json:"last_page,omitempty"`
}

// OpenAlexOpenAccess represents the open access status of a work.
type OpenAlexOpenAccess struct {
	IsOA               bool   `json:"is_oa"`
	OAStatus           string `json:"oa_status,omitempty"`
	OAURL              string `json:"oa_url,omitempty"`
	RepositoryFulltext bool   `json:"any_repository_has_fulltext"`
}

// OpenAlexLocation represents a location where a work is hosted.
type OpenAlexLocation struct {
	LandingPageURL string         `json:"landing_page_url,omitempty"`
	PDFURL         string         `json:"pdf_url,omitempty"`
	License        string         `json:"license,omitempty"`
	Version        string         `json:"version,omitempty"`
	Source         OpenAlexSource `json:"source,omitempty"`
}

// OpenAlexSource represents the source (journal/repository) of a location.
type OpenAlexSource struct {
	OpenAlexID  string `json:"openalex_id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	ISSN        string `json:"issn,omitempty"`
}

// OpenAlexAuthorship represents an author with institutional affiliations.
type OpenAlexAuthorship struct {
	Position           string                `json:"position,omitempty"`
	Author             OpenAlexAuthor        `json:"author"`
	Institutions       []OpenAlexInstitution `json:"institutions,omitempty"`
	AffiliationStrings []string              `json:"affiliation_strings,omitempty"`
}

// OpenAlexAuthor represents an author entity with verified ORCID.
type OpenAlexAuthor struct {
	OpenAlexID  string `json:"openalex_id,omitempty"`
	ORCID       string `json:"orcid,omitempty"`
	DisplayName string `json:"display_name"`
}

// OpenAlexInstitution represents an institution with its ROR identifier.
type OpenAlexInstitution struct {
	OpenAlexID  string `json:"openalex_id,omitempty"`
	ROR         string `json:"ror,omitempty"`
	DisplayName string `json:"display_name"`
	CountryCode string `json:"country_code,omitempty"`
}

// OpenAlexTopic represents a hierarchical topic classification
// (Domain -> Field -> Subfield -> Topic).
type OpenAlexTopic struct {
	OpenAlexID  string        `json:"openalex_id,omitempty"`
	DisplayName string        `json:"display_name"`
	Score       float64       `json:"score"`
	Subfield    OpenAlexTaxon `json:"subfield,omitempty"`
	Field       OpenAlexTaxon `json:"field,omitempty"`
	Domain      OpenAlexTaxon `json:"domain,omitempty"`
}

// OpenAlexTaxon represents a node in the OpenAlex topic taxonomy.
type OpenAlexTaxon struct {
	OpenAlexID  string `json:"openalex_id,omitempty"`
	DisplayName string `json:"display_name"`
}

// OpenAlexKeyword represents a keyword attached to a work.
type OpenAlexKeyword struct {
	OpenAlexID  string  `json:"openalex_id,omitempty"`
	DisplayName string  `json:"display_name"`
	Score       float64 `json:"score"`
}

// OpenAlexAward represents a grant/funding acknowledgment.
type OpenAlexAward struct {
	Funder      OpenAlexFunder `json:"funder"`
	AwardID     string         `json:"award_id,omitempty"`
	AwardAmount int            `json:"award_amount,omitempty"`
	FundingYear int            `json:"funding_year,omitempty"`
}

// OpenAlexFunder represents a funding entity identified by its OpenAlex ID.
type OpenAlexFunder struct {
	OpenAlexID  string `json:"openalex_id,omitempty"`
	DisplayName string `json:"display_name"`
}

// CitationMetrics holds citation impact metrics for a work.
type CitationMetrics struct {
	CitedByCount  int     `json:"cited_by_count"`
	FWCI          float64 `json:"fwci"`
	PercentileMin float64 `json:"percentile_min"`
	PercentileMax float64 `json:"percentile_max"`
}

// OpenAlexWorksResult holds a page of works from a list request.
type OpenAlexWorksResult struct {
	TotalCount int             `json:"total_count"`
	PerPage    int             `json:"per_page"`
	Cursor     string          `json:"cursor"`
	Results    []*OpenAlexWork `json:"results"`
}
