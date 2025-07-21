package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// EuropePMCService handles API interactions with EuropePMC.
type EuropePMCService struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	email      string
}

// EuropePMCServiceOption configures the EuropePMC service.
type EuropePMCServiceOption func(*EuropePMCService)

// WithEuropePMCHTTPClient sets the HTTP client for the service.
func WithEuropePMCHTTPClient(client *http.Client) EuropePMCServiceOption {
	return func(s *EuropePMCService) {
		if client != nil {
			s.httpClient = client
		}
	}
}

// WithEuropePMCUserAgent sets the User-Agent header.
func WithEuropePMCUserAgent(userAgent string) EuropePMCServiceOption {
	return func(s *EuropePMCService) {
		if userAgent != "" {
			s.userAgent = userAgent
		}
	}
}

// WithEuropePMCEmail sets the email contact information.
func WithEuropePMCEmail(email string) EuropePMCServiceOption {
	return func(s *EuropePMCService) {
		if email != "" {
			s.email = email
		}
	}
}

// NewEuropePMCService creates a new EuropePMC service with the provided options.
func NewEuropePMCService(opts ...EuropePMCServiceOption) *EuropePMCService {
	service := &EuropePMCService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://www.ebi.ac.uk/europepmc/webservices/rest",
		userAgent:  "europepmc-go-client/1.0",
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

// FetchArticle retrieves article metadata for the given PMID from EuropePMC.
func (s *EuropePMCService) FetchArticle(
	pmid string,
) (*EuropePMCAPIResponse, error) {
	query := fmt.Sprintf("ext_id:%s", pmid)
	params := EuropePMCSearchParams{
		Query:      query,
		ResultType: "core",
		Format:     "json",
		PageSize:   1,
		CursorMark: "*",
	}

	return s.SearchArticles(params)
}

// SearchArticles performs a search query against EuropePMC API.
func (s *EuropePMCService) SearchArticles(
	params EuropePMCSearchParams,
) (*EuropePMCAPIResponse, error) {
	searchURL, err := s.buildSearchURL(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")
	if s.email != "" {
		req.Header.Set("X-Contact-Email", s.email)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"API request failed with status %d",
			resp.StatusCode,
		)
	}

	apiResponse := &EuropePMCAPIResponse{}
	if err := json.NewDecoder(resp.Body).Decode(apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResponse, nil
}

// buildSearchURL constructs the search URL with query parameters.
func (s *EuropePMCService) buildSearchURL(
	params EuropePMCSearchParams,
) (string, error) {
	baseURL, err := url.Parse(s.baseURL + "/search")
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	queryParams := url.Values{}
	queryParams.Set("query", params.Query)

	if params.ResultType != "" {
		queryParams.Set("resultType", params.ResultType)
	}
	if params.Format != "" {
		queryParams.Set("format", params.Format)
	}
	if params.PageSize > 0 {
		queryParams.Set("pageSize", strconv.Itoa(params.PageSize))
	}
	if params.CursorMark != "" {
		queryParams.Set("cursorMark", params.CursorMark)
	}
	if params.Sort != "" {
		queryParams.Set("sort", params.Sort)
	}

	baseURL.RawQuery = queryParams.Encode()
	return baseURL.String(), nil
}

// EuropePMCSearchParams holds parameters for EuropePMC search requests.
type EuropePMCSearchParams struct {
	Query      string
	ResultType string
	Format     string
	PageSize   int
	CursorMark string
	Sort       string
}

// EuropePMCAPIResponse represents the raw API response from EuropePMC.
type EuropePMCAPIResponse struct {
	Version    string                 `json:"version"`
	HitCount   int                    `json:"hitCount"`
	Request    EuropePMCAPIRequest    `json:"request"`
	ResultList EuropePMCAPIResultList `json:"resultList"`
}

// EuropePMCAPIRequest represents the request information echoed back by the API.
type EuropePMCAPIRequest struct {
	QueryString string `json:"queryString"`
	ResultType  string `json:"resultType"`
	CursorMark  string `json:"cursorMark"`
	PageSize    int    `json:"pageSize"`
	Sort        string `json:"sort"`
	Synonym     bool   `json:"synonym"`
}

// EuropePMCAPIResultList contains the list of results from the API.
type EuropePMCAPIResultList struct {
	Result []EuropePMCAPIArticle `json:"result"`
}

// EuropePMCAPIArticle represents the raw article data from EuropePMC API.
type EuropePMCAPIArticle struct {
	ID                    string                     `json:"id"`
	Source                string                     `json:"source"`
	PMID                  string                     `json:"pmid"`
	PMCID                 string                     `json:"pmcid"`
	FullTextIdList        EuropePMCFullTextIdList    `json:"fullTextIdList"`
	DOI                   string                     `json:"doi"`
	Title                 string                     `json:"title"`
	AuthorString          string                     `json:"authorString"`
	AuthorList            EuropePMCAPIAuthorList     `json:"authorList"`
	AuthorIdList          EuropePMCAPIAuthorIdList   `json:"authorIdList"`
	DataLinksTagsList     EuropePMCDataLinksTagsList `json:"dataLinksTagsList"`
	JournalInfo           EuropePMCAPIJournalInfo    `json:"journalInfo"`
	PubYear               string                     `json:"pubYear"`
	PageInfo              string                     `json:"pageInfo"`
	AbstractText          string                     `json:"abstractText"`
	Affiliation           string                     `json:"affiliation"`
	PublicationStatus     string                     `json:"publicationStatus"`
	Language              string                     `json:"language"`
	PubModel              string                     `json:"pubModel"`
	PubTypeList           EuropePMCPubTypeList       `json:"pubTypeList"`
	GrantsList            EuropePMCGrantsList        `json:"grantsList"`
	MeshHeadingList       EuropePMCMeshHeadingList   `json:"meshHeadingList"`
	KeywordList           EuropePMCKeywordList       `json:"keywordList"`
	ChemicalList          EuropePMCChemicalList      `json:"chemicalList"`
	SubsetList            EuropePMCSubsetList        `json:"subsetList"`
	FullTextURLList       EuropePMCFullTextURLList   `json:"fullTextUrlList"`
	IsOpenAccess          string                     `json:"isOpenAccess"`
	InEPMC                string                     `json:"inEPMC"`
	InPMC                 string                     `json:"inPMC"`
	HasPDF                string                     `json:"hasPDF"`
	HasBook               string                     `json:"hasBook"`
	HasSuppl              string                     `json:"hasSuppl"`
	CitedByCount          int                        `json:"citedByCount"`
	HasData               string                     `json:"hasData"`
	HasReferences         string                     `json:"hasReferences"`
	HasTextMinedTerms     string                     `json:"hasTextMinedTerms"`
	HasDbCrossReferences  string                     `json:"hasDbCrossReferences"`
	HasLabsLinks          string                     `json:"hasLabsLinks"`
	License               string                     `json:"license"`
	HasEvaluations        string                     `json:"hasEvaluations"`
	AuthMan               string                     `json:"authMan"`
	EpmcAuthMan           string                     `json:"epmcAuthMan"`
	NihAuthMan            string                     `json:"nihAuthMan"`
	HasTMAccessionNumbers string                     `json:"hasTMAccessionNumbers"`
	DateOfCompletion      string                     `json:"dateOfCompletion"`
	DateOfCreation        string                     `json:"dateOfCreation"`
	FirstIndexDate        string                     `json:"firstIndexDate"`
	FullTextReceivedDate  string                     `json:"fullTextReceivedDate"`
	DateOfRevision        string                     `json:"dateOfRevision"`
	FirstPublicationDate  string                     `json:"firstPublicationDate"`
}

// Supporting API types for nested structures.
type EuropePMCFullTextIdList struct {
	FullTextId []string `json:"fullTextId"`
}

type EuropePMCAPIAuthorList struct {
	Author []EuropePMCAPIAuthor `json:"author"`
}

type EuropePMCAPIAuthor struct {
	FullName                     string                            `json:"fullName"`
	FirstName                    string                            `json:"firstName"`
	LastName                     string                            `json:"lastName"`
	Initials                     string                            `json:"initials"`
	AuthorId                     EuropePMCAPIAuthorId              `json:"authorId"`
	AuthorAffiliationDetailsList EuropePMCAPIAuthorAffiliationList `json:"authorAffiliationDetailsList"`
}

type EuropePMCAPIAuthorId struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type EuropePMCAPIAuthorAffiliationList struct {
	AuthorAffiliation []EuropePMCAPIAuthorAffiliation `json:"authorAffiliation"`
}

type EuropePMCAPIAuthorAffiliation struct {
	Affiliation string `json:"affiliation"`
}

type EuropePMCAPIAuthorIdList struct {
	AuthorId []EuropePMCAPIAuthorId `json:"authorId"`
}

type EuropePMCDataLinksTagsList struct {
	DataLinkstag []string `json:"dataLinkstag"`
}

type EuropePMCAPIJournalInfo struct {
	Issue                string              `json:"issue"`
	Volume               string              `json:"volume"`
	JournalIssueId       int                 `json:"journalIssueId"`
	DateOfPublication    string              `json:"dateOfPublication"`
	MonthOfPublication   int                 `json:"monthOfPublication"`
	YearOfPublication    int                 `json:"yearOfPublication"`
	PrintPublicationDate string              `json:"printPublicationDate"`
	Journal              EuropePMCAPIJournal `json:"journal"`
}

type EuropePMCAPIJournal struct {
	Title               string `json:"title"`
	MedlineAbbreviation string `json:"medlineAbbreviation"`
	ISSN                string `json:"issn"`
	ESSN                string `json:"essn"`
	IsoAbbreviation     string `json:"isoabbreviation"`
	NLMID               string `json:"nlmid"`
}

type EuropePMCPubTypeList struct {
	PubType []string `json:"pubType"`
}

type EuropePMCGrantsList struct {
	Grant []EuropePMCAPIGrant `json:"grant"`
}

type EuropePMCAPIGrant struct {
	GrantId string `json:"grantId"`
	Agency  string `json:"agency"`
	OrderIn int    `json:"orderIn"`
}

type EuropePMCMeshHeadingList struct {
	MeshHeading []EuropePMCAPIMeshHeading `json:"meshHeading"`
}

type EuropePMCAPIMeshHeading struct {
	MajorTopicYN      string                     `json:"majorTopic_YN"`
	DescriptorName    string                     `json:"descriptorName"`
	MeshQualifierList EuropePMCMeshQualifierList `json:"meshQualifierList"`
}

type EuropePMCMeshQualifierList struct {
	MeshQualifier []EuropePMCAPIMeshQualifier `json:"meshQualifier"`
}

type EuropePMCAPIMeshQualifier struct {
	Abbreviation  string `json:"abbreviation"`
	QualifierName string `json:"qualifierName"`
	MajorTopicYN  string `json:"majorTopic_YN"`
}

type EuropePMCKeywordList struct {
	Keyword []string `json:"keyword"`
}

type EuropePMCChemicalList struct {
	Chemical []EuropePMCAPIChemical `json:"chemical"`
}

type EuropePMCAPIChemical struct {
	Name           string `json:"name"`
	RegistryNumber string `json:"registryNumber"`
}

type EuropePMCSubsetList struct {
	Subset []EuropePMCAPISubset `json:"subset"`
}

type EuropePMCAPISubset struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type EuropePMCFullTextURLList struct {
	FullTextURL []EuropePMCAPIFullTextURL `json:"fullTextUrl"`
}

type EuropePMCAPIFullTextURL struct {
	Availability     string `json:"availability"`
	AvailabilityCode string `json:"availabilityCode"`
	DocumentStyle    string `json:"documentStyle"`
	Site             string `json:"site"`
	URL              string `json:"url"`
}
