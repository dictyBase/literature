package literature

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dictybase/literature/internal"
	"github.com/go-playground/validator/v10"
)

// pmidPrefix is the identifier prefix OpenAlex uses for PubMed IDs.
const pmidPrefix = "PMID:"

// filterChunkSize bounds how many OR-separated filter values are sent in
// a single works list request (the OpenAlex OR-filter limit).
const filterChunkSize = 50

// OpenAlexClient provides access to OpenAlex scholarly metadata, citation
// metrics, and citation networks.
type OpenAlexClient struct {
	openAlexService *internal.OpenAlexService
	httpClient      *http.Client
	baseURL         string
	userAgent       string
	email           string
	apiKey          string
	validate        *validator.Validate
}

// NewOpenAlexClient creates a new OpenAlex client with the provided options.
func NewOpenAlexClient(opts ...OpenAlexOption) (*OpenAlexClient, error) {
	client := &OpenAlexClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.openalex.org",
		userAgent:  "openalex-go-client/1.0",
		validate:   validator.New(),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Initialize services using internal constructors with appropriate options
	var serviceOpts []internal.OpenAlexServiceOption

	if client.httpClient != nil {
		serviceOpts = append(
			serviceOpts,
			internal.WithOpenAlexHTTPClient(client.httpClient),
		)
	}
	if client.userAgent != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithOpenAlexUserAgent(client.userAgent),
		)
	}
	if client.email != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithOpenAlexEmail(client.email),
		)
	}
	if client.apiKey != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithOpenAlexAPIKey(client.apiKey),
		)
	}
	if client.baseURL != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithOpenAlexBaseURL(client.baseURL),
		)
	}

	client.openAlexService = internal.NewOpenAlexService(serviceOpts...)

	return client, nil
}

// GetWork retrieves a single work by its OpenAlex ID (W...), a PMID
// (PMID:12345678 or bare number), or a DOI.
func (c *OpenAlexClient) GetWork(workID string) (*OpenAlexWork, error) {
	if err := c.validate.Var(workID, "required"); err != nil {
		return nil, invalidInputError(err)
	}

	apiWork, err := c.openAlexService.FetchWork(
		stripOpenAlexIDPrefix(workID),
	)
	if err != nil {
		return nil, c.createWorkError(workID, err)
	}

	return convertFromInternalOpenAlexWork(apiWork), nil
}

// GetWorkByPMID retrieves a single work by its PubMed ID.
func (c *OpenAlexClient) GetWorkByPMID(pmid string) (*OpenAlexWork, error) {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return nil, invalidInputError(err)
	}

	return c.GetWork(normalizePMID(pmid))
}

// GetWorkByDOI retrieves a single work by its DOI. Leading "doi:" and
// "https://doi.org/" prefixes are accepted.
func (c *OpenAlexClient) GetWorkByDOI(doi string) (*OpenAlexWork, error) {
	if err := c.validate.Var(doi, "required"); err != nil {
		return nil, invalidInputError(err)
	}

	return c.GetWork(doi)
}

// GetWorks retrieves multiple works using their OpenAlex IDs. Identifiers
// are sent in chunked requests bounded by the OpenAlex OR-filter limit.
func (c *OpenAlexClient) GetWorks(ids []string) ([]*OpenAlexWork, error) {
	if err := c.validate.Var(ids, "required,min=1,dive,required"); err != nil {
		return nil, invalidInputError(err)
	}

	return c.fetchWorksByFilter(
		"ids.openalex",
		ids,
		stripOpenAlexIDPrefix,
	)
}

// GetWorksByPMIDs retrieves multiple works using their PubMed IDs.
// Identifiers are sent in chunked requests bounded by the OpenAlex
// OR-filter limit.
func (c *OpenAlexClient) GetWorksByPMIDs(
	pmids []string,
) ([]*OpenAlexWork, error) {
	if err := c.validate.Var(pmids, "required,min=1,dive,required"); err != nil {
		return nil, invalidInputError(err)
	}

	return c.fetchWorksByFilter(
		"pmid",
		pmids,
		normalizePMID,
	)
}

// GetReferencedWorks resolves all works cited by the given work. The
// referenced IDs are batch-resolved in a single request.
func (c *OpenAlexClient) GetReferencedWorks(
	workID string,
) ([]*OpenAlexWork, error) {
	if err := c.validate.Var(workID, "required"); err != nil {
		return nil, invalidInputError(err)
	}

	work, err := c.GetWork(workID)
	if err != nil {
		return nil, err
	}

	if len(work.ReferencedWorks) == 0 {
		return []*OpenAlexWork{}, nil
	}

	return c.GetWorks(work.ReferencedWorks)
}

// GetCitingWorks retrieves a single page of works that cite the given
// work using the filter=cites:<id> query. Page size, cursor, and sort are
// controlled with WithOpenAlexPerPage, WithOpenAlexCursor, and
// WithOpenAlexSort; use the returned Cursor to fetch subsequent pages.
func (c *OpenAlexClient) GetCitingWorks(
	workID string,
	opts ...OpenAlexSearchOption,
) (*OpenAlexWorksResult, error) {
	if err := c.validate.Var(workID, "required"); err != nil {
		return nil, invalidInputError(err)
	}

	config := &openAlexSearchConfig{}
	for _, opt := range opts {
		opt(config)
	}

	params := internal.OpenAlexWorksParams{
		Filter:  fmt.Sprintf("cites:%s", workID),
		PerPage: config.perPage,
		Cursor:  config.cursor,
		Sort:    config.sort,
	}

	apiResponse, err := c.openAlexService.SearchWorks(params)
	if err != nil {
		return nil, c.createWorkError(workID, err)
	}

	return convertFromInternalOpenAlexWorksResult(apiResponse), nil
}

// GetCitationMetrics returns citation count, Field-Weighted Citation
// Impact (FWCI), and normalized citation percentiles for a work.
func (c *OpenAlexClient) GetCitationMetrics(
	workID string,
) (*CitationMetrics, error) {
	if err := c.validate.Var(workID, "required"); err != nil {
		return nil, invalidInputError(err)
	}

	work, err := c.GetWork(workID)
	if err != nil {
		return nil, err
	}

	return &CitationMetrics{
		CitedByCount:  work.CitedByCount,
		FWCI:          work.FWCI,
		PercentileMin: work.PercentileMin,
		PercentileMax: work.PercentileMax,
	}, nil
}

// fetchWorksByFilter performs batch works lookups keyed by the given
// filter with identifier transformation. Identifiers are chunked into
// requests bounded by the OpenAlex OR-filter limit, and each request asks
// for exactly the chunk size so results are never silently truncated by
// the default 25-item page.
func (c *OpenAlexClient) fetchWorksByFilter(
	filterKey string,
	identifiers []string,
	normalize func(string) string,
) ([]*OpenAlexWork, error) {
	normalized := make([]string, len(identifiers))
	for i, identifier := range identifiers {
		normalized[i] = normalize(identifier)
	}

	works := make([]*OpenAlexWork, 0, len(normalized))
	for start := 0; start < len(normalized); start += filterChunkSize {
		end := min(start+filterChunkSize, len(normalized))
		chunk := normalized[start:end]

		filter := fmt.Sprintf("%s:%s", filterKey, strings.Join(chunk, "|"))
		params := internal.OpenAlexWorksParams{
			Filter:  filter,
			PerPage: len(chunk),
		}

		apiResponse, err := c.openAlexService.SearchWorks(params)
		if err != nil {
			return nil, &Error{
				Type:    ErrorTypeAPIError,
				Message: fmt.Sprintf("batch lookup failed: %s", err.Error()),
				Query:   filter,
				Cause:   err,
			}
		}

		for i := range apiResponse.Results {
			works = append(
				works,
				convertFromInternalOpenAlexWork(&apiResponse.Results[i]),
			)
		}
	}

	return works, nil
}

// createWorkError maps a service error to a public client error with
// identifier context.
func (c *OpenAlexClient) createWorkError(
	identifier string,
	cause error,
) *Error {
	errWithContext := &Error{
		Message: fmt.Sprintf("failed to fetch work: %s", cause.Error()),
		Cause:   cause,
	}

	switch {
	case errors.Is(cause, internal.ErrOpenAlexNotFound):
		errWithContext.Type = ErrorTypeArticleNotFound
		errWithContext.Message = "work not found"
	case errors.Is(cause, internal.ErrOpenAlexRateLimited):
		errWithContext.Type = ErrorTypeRateLimit
		errWithContext.Message = "rate limit exceeded"
	default:
		errWithContext.Type = ErrorTypeAPIError
	}

	c.setErrorIdentifier(errWithContext, identifier)

	return errWithContext
}

// setErrorIdentifier associates the identifier with the error using its
// detected type (PMID, DOI, or generic query identifier).
func (c *OpenAlexClient) setErrorIdentifier(err *Error, identifier string) {
	switch {
	case isPMIDIdentifier(identifier):
		err.PMID = identifier
	case isDOIIdentifier(identifier):
		err.DOI = identifier
	default:
		err.Query = identifier
	}
}

// normalizePMID ensures the PMID carries the PMID: prefix OpenAlex expects.
func normalizePMID(pmid string) string {
	if isPMIDIdentifier(pmid) {
		return pmid
	}
	return pmidPrefix + pmid
}

// isPMIDIdentifier reports whether the identifier is already PMID-prefixed.
func isPMIDIdentifier(identifier string) bool {
	return strings.HasPrefix(strings.ToUpper(identifier), pmidPrefix)
}

// isDOIIdentifier reports whether the identifier refers to a DOI.
func isDOIIdentifier(identifier string) bool {
	lower := strings.ToLower(identifier)
	if strings.HasPrefix(lower, "doi:") ||
		strings.HasPrefix(lower, "https://doi.org/") {
		return true
	}
	return strings.HasPrefix(identifier, "10.") &&
		strings.Contains(identifier, "/")
}

// invalidInputError creates a validation failure error.
func invalidInputError(err error) *Error {
	return &Error{
		Type:    ErrorTypeInvalidInput,
		Message: fmt.Sprintf("validation failed: %s", err.Error()),
	}
}

// convertFromInternalOpenAlexWork maps a raw API work to the public type.
func convertFromInternalOpenAlexWork(
	apiWork *internal.OpenAlexAPIWork,
) *OpenAlexWork {
	work := &OpenAlexWork{
		OpenAlexID:      stripOpenAlexIDPrefix(apiWork.ID),
		DOI:             stripDOIPrefix(apiWork.DOI),
		PMID:            stripPMIDPrefix(apiWork.OpenAlexIDs.PMID),
		Title:           apiWork.DisplayName,
		PublicationYear: apiWork.PublicationYear,
		PublicationDate: apiWork.PublicationDate,
		Type:            apiWork.Type,
		CitedByCount:    apiWork.CitedByCount,
		FWCI:            apiWork.FWCI,
		PercentileMin:   apiWork.CitedByPercentileYear.Min,
		PercentileMax:   apiWork.CitedByPercentileYear.Max,
		ReferencedWorks: apiWork.ReferencedWorks,
		RelatedWorks:    apiWork.RelatedWorks,
		IsRetracted:     apiWork.IsRetracted,
		OpenAccess:      convertFromInternalOpenAccess(apiWork.OpenAccess),
	}
	work.PrimaryLocation = *convertFromInternalOpenAlexLocation(
		apiWork.PrimaryLocation,
	)
	work.Authors = convertFromInternalOpenAlexAuthorships(apiWork.Authorships)
	work.Topics = convertFromInternalOpenAlexTopics(apiWork.Topics)

	if apiWork.BestOALocation != nil {
		work.BestOALocation = convertFromInternalOpenAlexLocation(
			*apiWork.BestOALocation,
		)
	}

	work.Keywords = make([]OpenAlexKeyword, len(apiWork.Keywords))
	for i, keyword := range apiWork.Keywords {
		work.Keywords[i] = OpenAlexKeyword{
			OpenAlexID:  stripOpenAlexIDPrefix(keyword.ID),
			DisplayName: keyword.DisplayName,
			Score:       keyword.Score,
		}
	}

	work.Awards = make([]OpenAlexAward, len(apiWork.Grants))
	for i, grant := range apiWork.Grants {
		work.Awards[i] = OpenAlexAward{
			Funder: OpenAlexFunder{
				OpenAlexID:  stripOpenAlexIDPrefix(grant.FunderID),
				DisplayName: grant.FunderDisplayName,
			},
			AwardID:     grant.AwardID,
			AwardAmount: grant.AwardAmount,
			FundingYear: grant.FundingYear,
		}
	}

	work.Volume = apiWork.Biblio.Volume
	work.Issue = apiWork.Biblio.Issue
	work.FirstPage = apiWork.Biblio.FirstPage
	work.LastPage = apiWork.Biblio.LastPage

	return work
}

// convertFromInternalOpenAlexWorksResult maps a raw works list response to
// the public result type.
func convertFromInternalOpenAlexWorksResult(
	apiResponse *internal.OpenAlexAPIWorksResponse,
) *OpenAlexWorksResult {
	result := &OpenAlexWorksResult{
		TotalCount: apiResponse.Meta.Count,
		PerPage:    apiResponse.Meta.PerPage,
		Cursor:     apiResponse.Meta.Cursor,
		Results: make(
			[]*OpenAlexWork,
			len(apiResponse.Results),
		),
	}

	for i := range apiResponse.Results {
		result.Results[i] = convertFromInternalOpenAlexWork(
			&apiResponse.Results[i],
		)
	}

	return result
}

// convertFromInternalOpenAccess maps open access status.
func convertFromInternalOpenAccess(
	apiOpenAccess internal.OpenAlexAPIOpenAccess,
) OpenAlexOpenAccess {
	return OpenAlexOpenAccess{
		IsOA:               apiOpenAccess.IsOA,
		OAStatus:           apiOpenAccess.OAStatus,
		OAURL:              apiOpenAccess.OAURL,
		RepositoryFulltext: apiOpenAccess.AnyRepositoryHasFulltext,
	}
}

// convertFromInternalOpenAlexLocation maps a location.
func convertFromInternalOpenAlexLocation(
	apiLocation internal.OpenAlexAPILocation,
) *OpenAlexLocation {
	return &OpenAlexLocation{
		LandingPageURL: apiLocation.LandingPageURL,
		PDFURL:         apiLocation.PDFURL,
		License:        apiLocation.License,
		Version:        apiLocation.Version,
		Source: OpenAlexSource{
			OpenAlexID:  stripOpenAlexIDPrefix(apiLocation.Source.ID),
			DisplayName: apiLocation.Source.DisplayName,
			ISSN:        firstISSN(apiLocation.Source),
		},
	}
}

// convertFromInternalOpenAlexAuthorships maps authorships.
func convertFromInternalOpenAlexAuthorships(
	apiAuthorships []internal.OpenAlexAPIAuthorship,
) []OpenAlexAuthorship {
	authorships := make(
		[]OpenAlexAuthorship,
		len(apiAuthorships),
	)
	for idx, apiAuthorship := range apiAuthorships {
		institutions := make(
			[]OpenAlexInstitution,
			len(apiAuthorship.Institutions),
		)
		for j, apiInstitution := range apiAuthorship.Institutions {
			institutions[j] = OpenAlexInstitution{
				OpenAlexID:  stripOpenAlexIDPrefix(apiInstitution.ID),
				ROR:         stripRORPrefix(apiInstitution.ROR),
				DisplayName: apiInstitution.DisplayName,
				CountryCode: apiInstitution.CountryCode,
			}
		}
		authorships[idx] = OpenAlexAuthorship{
			Position: apiAuthorship.AuthorPosition,
			Author: OpenAlexAuthor{
				OpenAlexID:  stripOpenAlexIDPrefix(apiAuthorship.Author.ID),
				ORCID:       stripORCIDPrefix(apiAuthorship.Author.ORCID),
				DisplayName: apiAuthorship.Author.DisplayName,
			},
			Institutions:       institutions,
			AffiliationStrings: apiAuthorship.RawAffiliationStrings,
		}
	}

	return authorships
}

// convertFromInternalOpenAlexTopics maps hierarchical topics.
func convertFromInternalOpenAlexTopics(
	apiTopics []internal.OpenAlexAPITopic,
) []OpenAlexTopic {
	topics := make([]OpenAlexTopic, len(apiTopics))
	for i, apiTopic := range apiTopics {
		topics[i] = OpenAlexTopic{
			OpenAlexID:  stripOpenAlexIDPrefix(apiTopic.ID),
			DisplayName: apiTopic.DisplayName,
			Score:       apiTopic.Score,
			Subfield:    convertFromInternalOpenAlexTaxon(apiTopic.Subfield),
			Field:       convertFromInternalOpenAlexTaxon(apiTopic.Field),
			Domain:      convertFromInternalOpenAlexTaxon(apiTopic.Domain),
		}
	}

	return topics
}

// convertFromInternalOpenAlexTaxon maps a taxonomy node.
func convertFromInternalOpenAlexTaxon(
	apiTaxon internal.OpenAlexAPITopicNode,
) OpenAlexTaxon {
	return OpenAlexTaxon{
		OpenAlexID:  stripOpenAlexIDPrefix(apiTaxon.ID),
		DisplayName: apiTaxon.DisplayName,
	}
}

// firstISSN returns the ISSN-L when present, falling back to the first
// listed ISSN.
func firstISSN(source internal.OpenAlexAPISourceInfo) string {
	if source.ISSNL != "" {
		return source.ISSNL
	}
	if len(source.ISSN) > 0 {
		return source.ISSN[0]
	}
	return ""
}

// stripOpenAlexIDPrefix trims the entity prefix from an OpenAlex URL,
// e.g. https://openalex.org/W123 -> W123.
func stripOpenAlexIDPrefix(id string) string {
	const prefix = "https://openalex.org/"
	if strings.HasPrefix(id, prefix) {
		return strings.TrimPrefix(id, prefix)
	}
	return id
}

// stripDOIPrefix trims the https://doi.org/ prefix from a DOI URL.
func stripDOIPrefix(doi string) string {
	const prefix = "https://doi.org/"
	if strings.HasPrefix(doi, prefix) {
		return strings.TrimPrefix(doi, prefix)
	}
	return doi
}

// stripPMIDPrefix trims the PMID: prefix, e.g. PMID:123 -> 123.
func stripPMIDPrefix(pmid string) string {
	if strings.HasPrefix(pmid, pmidPrefix) {
		return strings.TrimPrefix(pmid, pmidPrefix)
	}
	return pmid
}

// stripRORPrefix trims the https://ror.org/ prefix from a ROR URL.
func stripRORPrefix(ror string) string {
	const prefix = "https://ror.org/"
	if strings.HasPrefix(ror, prefix) {
		return strings.TrimPrefix(ror, prefix)
	}
	return ror
}

// stripORCIDPrefix trims the https://orcid.org/ prefix from an ORCID URL.
func stripORCIDPrefix(orcid string) string {
	const prefix = "https://orcid.org/"
	if strings.HasPrefix(orcid, prefix) {
		return strings.TrimPrefix(orcid, prefix)
	}
	return orcid
}
