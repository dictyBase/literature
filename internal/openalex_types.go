package internal

// OpenAlexAPIWork represents the raw work object returned by the OpenAlex API.
type OpenAlexAPIWork struct {
	ID                    string                    `json:"id"`
	DOI                   string                    `json:"doi"`
	OpenAlexIDs           OpenAlexAPIIdentifiers    `json:"ids"`
	DisplayName           string                    `json:"display_name"`
	PublicationYear       int                       `json:"publication_year"`
	PublicationDate       string                    `json:"publication_date"`
	Type                  string                    `json:"type"`
	CitedByCount          int                       `json:"cited_by_count"`
	FWCI                  float64                   `json:"fwci"`
	CitedByPercentileYear OpenAlexAPIPercentileYear `json:"cited_by_percentile_year"`
	ReferencedWorks       []string                  `json:"referenced_works"`
	RelatedWorks          []string                  `json:"related_works"`
	IsRetracted           bool                      `json:"is_retracted"`
	PrimaryLocation       OpenAlexAPILocation       `json:"primary_location"`
	BestOALocation        *OpenAlexAPILocation      `json:"best_oa_location"`
	OpenAccess            OpenAlexAPIOpenAccess     `json:"open_access"`
	Authorships           []OpenAlexAPIAuthorship   `json:"authorships"`
	Topics                []OpenAlexAPITopic        `json:"topics"`
	Keywords              []OpenAlexAPIKeyword      `json:"keywords"`
	Grants                []OpenAlexAPIGrant        `json:"grants"`
	Biblio                OpenAlexAPIBiblio         `json:"biblio"`
}

// OpenAlexAPIIdentifiers holds the external identifiers of a work.
type OpenAlexAPIIdentifiers struct {
	OpenAlex string `json:"openalex"`
	DOI      string `json:"doi"`
	PMID     string `json:"pmid"`
	MAG      string `json:"mag"`
}

// OpenAlexAPIPercentileYear holds the normalized citation percentiles.
type OpenAlexAPIPercentileYear struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// OpenAlexAPILocation represents a location where a work is hosted.
type OpenAlexAPILocation struct {
	LandingPageURL string                `json:"landing_page_url"`
	PDFURL         string                `json:"pdf_url"`
	License        string                `json:"license"`
	Version        string                `json:"version"`
	Source         OpenAlexAPISourceInfo `json:"source"`
}

// OpenAlexAPISourceInfo represents the source (journal/repo) of a location.
type OpenAlexAPISourceInfo struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name"`
	ISSNL       string   `json:"issn_l"`
	ISSN        []string `json:"issn"`
}

// OpenAlexAPIOpenAccess represents the open access status of a work.
type OpenAlexAPIOpenAccess struct {
	IsOA                     bool   `json:"is_oa"`
	OAStatus                 string `json:"oa_status"`
	OAURL                    string `json:"oa_url"`
	AnyRepositoryHasFulltext bool   `json:"any_repository_has_fulltext"`
}

// OpenAlexAPIAuthorship represents an author with affiliations.
type OpenAlexAPIAuthorship struct {
	AuthorPosition        string                   `json:"author_position"`
	Author                OpenAlexAPIAuthor        `json:"author"`
	Institutions          []OpenAlexAPIInstitution `json:"institutions"`
	RawAffiliationStrings []string                 `json:"raw_affiliation_strings"`
}

// OpenAlexAPIAuthor represents an author entity.
type OpenAlexAPIAuthor struct {
	ID          string `json:"id"`
	ORCID       string `json:"orcid"`
	DisplayName string `json:"display_name"`
}

// OpenAlexAPIInstitution represents an institution entity.
type OpenAlexAPIInstitution struct {
	ID          string `json:"id"`
	ROR         string `json:"ror"`
	DisplayName string `json:"display_name"`
	CountryCode string `json:"country_code"`
	Type        string `json:"type"`
}

// OpenAlexAPITopic represents a hierarchical topic classification.
type OpenAlexAPITopic struct {
	ID          string               `json:"id"`
	DisplayName string               `json:"display_name"`
	Score       float64              `json:"score"`
	Subfield    OpenAlexAPITopicNode `json:"subfield"`
	Field       OpenAlexAPITopicNode `json:"field"`
	Domain      OpenAlexAPITopicNode `json:"domain"`
}

// OpenAlexAPITopicNode represents a node in the topic taxonomy.
type OpenAlexAPITopicNode struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// OpenAlexAPIKeyword represents a keyword attached to a work.
type OpenAlexAPIKeyword struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Score       float64 `json:"score"`
}

// OpenAlexAPIGrant represents a grant/funding acknowledgment.
type OpenAlexAPIGrant struct {
	AwardID           string `json:"award_id"`
	AwardAmount       int    `json:"award_amount"`
	AwardAmountUSD    int    `json:"award_amount_usd"`
	FunderID          string `json:"funder"`
	FunderDisplayName string `json:"funder_display_name"`
	FundingTimeSpan   string `json:"funding_time_span"`
	FundingYear       int    `json:"funding_year"`
}

// OpenAlexAPIBiblio holds bibliographic volume/issue/page info.
type OpenAlexAPIBiblio struct {
	Volume    string `json:"volume"`
	Issue     string `json:"issue"`
	FirstPage string `json:"first_page"`
	LastPage  string `json:"last_page"`
}

// OpenAlexAPIMeta holds pagination metadata of a list response.
type OpenAlexAPIMeta struct {
	Count   int    `json:"count"`
	PerPage int    `json:"per_page"`
	Cursor  string `json:"cursor"`
}

// OpenAlexAPIWorksResponse represents a list of works from the OpenAlex API.
type OpenAlexAPIWorksResponse struct {
	Meta    OpenAlexAPIMeta   `json:"meta"`
	Results []OpenAlexAPIWork `json:"results"`
}
