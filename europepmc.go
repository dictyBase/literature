package literature

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dictybase/literature/internal"
	"github.com/go-playground/validator/v10"
)

const (
	// identifierTypePMID represents PMID identifier type.
	identifierTypePMID = "PMID"
	// identifierTypeDOI represents DOI identifier type.
	identifierTypeDOI = "DOI"
)

// EuropePMCClient provides access to EuropePMC literature services.
type EuropePMCClient struct {
	europePMCService  *internal.EuropePMCService
	httpClient        *http.Client
	baseURL           string
	userAgent         string
	email             string
	maxRetries        int
	retryDelay        time.Duration
	requestsPerSecond float64
	defaultResultType string
	defaultFormat     string
	validate          *validator.Validate
}

// NewEuropePMCClient creates a new EuropePMC client with the provided options.
func NewEuropePMCClient(opts ...EuropePMCOption) (*EuropePMCClient, error) {
	client := &EuropePMCClient{
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		baseURL:           "https://www.ebi.ac.uk/europepmc/webservices/rest",
		userAgent:         "europepmc-go-client/1.0",
		maxRetries:        3,
		retryDelay:        time.Second,
		requestsPerSecond: 10.0,
		defaultResultType: "core",
		defaultFormat:     "json",
		validate:          validator.New(),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Initialize services using internal constructors with appropriate options
	var serviceOpts []internal.EuropePMCServiceOption

	// Pass HTTP client to internal service if configured
	if client.httpClient != nil {
		serviceOpts = append(
			serviceOpts,
			internal.WithEuropePMCHTTPClient(client.httpClient),
		)
	}
	if client.userAgent != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithEuropePMCUserAgent(client.userAgent),
		)
	}
	if client.email != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithEuropePMCEmail(client.email),
		)
	}
	if client.baseURL != "" {
		serviceOpts = append(
			serviceOpts,
			internal.WithEuropePMCBaseURL(client.baseURL),
		)
	}

	client.europePMCService = internal.NewEuropePMCService(serviceOpts...)

	return client, nil
}

// GetArticle retrieves article metadata for the given PMID from EuropePMC.
func (c *EuropePMCClient) GetArticle(pmid string) (*EuropePMCArticle, error) {
	return c.fetchArticleWithValidation(pmid, identifierTypePMID, func() (*internal.EuropePMCAPIResponse, error) {
		return c.europePMCService.FetchArticle(pmid)
	})
}

// GetArticleByDOI retrieves article metadata for the given DOI from EuropePMC.
func (c *EuropePMCClient) GetArticleByDOI(doi string) (*EuropePMCArticle, error) {
	return c.fetchArticleWithValidation(doi, identifierTypeDOI, func() (*internal.EuropePMCAPIResponse, error) {
		return c.europePMCService.FetchArticleByDOI(doi)
	})
}

// fetchArticleWithValidation is a helper function that handles the common logic
// for fetching articles with validation and error handling.
func (c *EuropePMCClient) fetchArticleWithValidation(
	identifier, identifierType string,
	fetchFunc func() (*internal.EuropePMCAPIResponse, error),
) (*EuropePMCArticle, error) {
	if err := c.validate.Var(identifier, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	apiResponse, err := fetchFunc()
	if err != nil {
		return nil, c.createAPIError(identifier, identifierType, err)
	}

	if apiResponse.HitCount == 0 {
		return nil, c.createNotFoundError(identifier, identifierType)
	}

	if len(apiResponse.ResultList.Result) == 0 {
		return nil, c.createParseError(identifier, identifierType)
	}

	return convertFromInternalEuropePMCArticle(
		&apiResponse.ResultList.Result[0],
	), nil
}

// createAPIError creates an API error with the appropriate identifier context.
func (c *EuropePMCClient) createAPIError(identifier, identifierType string, cause error) *Error {
	errWithContext := &Error{
		Type:    ErrorTypeAPIError,
		Message: fmt.Sprintf("failed to fetch article: %s", cause.Error()),
		Cause:   cause,
	}
	c.setErrorIdentifier(errWithContext, identifier, identifierType)
	return errWithContext
}

// createNotFoundError creates a not found error with the appropriate identifier context.
func (c *EuropePMCClient) createNotFoundError(identifier, identifierType string) *Error {
	errWithContext := &Error{
		Type:    ErrorTypeArticleNotFound,
		Message: "article not found",
	}
	c.setErrorIdentifier(errWithContext, identifier, identifierType)
	return errWithContext
}

// createParseError creates a parse error with the appropriate identifier context.
func (c *EuropePMCClient) createParseError(identifier, identifierType string) *Error {
	errWithContext := &Error{
		Type:    ErrorTypeParseError,
		Message: "no results in response",
	}
	c.setErrorIdentifier(errWithContext, identifier, identifierType)
	return errWithContext
}

// setErrorIdentifier sets the appropriate identifier field on an error based on type.
func (c *EuropePMCClient) setErrorIdentifier(err *Error, identifier, identifierType string) {
	switch identifierType {
	case identifierTypePMID:
		err.PMID = identifier
	case identifierTypeDOI:
		err.DOI = identifier
	}
}

// fetchArticleWithError is a helper function that wraps GetArticle for
// functional composition.
func (c *EuropePMCClient) fetchArticleWithError(
	pmid string,
) (*EuropePMCArticle, error) {
	article, err := c.GetArticle(pmid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article %s: %w", pmid, err)
	}
	return article, nil
}

// GetArticles retrieves metadata for multiple PMIDs from EuropePMC.
func (c *EuropePMCClient) GetArticles(
	pmids []string,
) ([]*EuropePMCArticle, error) {
	if err := c.validate.Var(pmids, "required,min=1,dive,required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	return internal.MapWithError(pmids, c.fetchArticleWithError)
}

// Search performs a literature search with the given query against EuropePMC.
func (c *EuropePMCClient) Search(
	query string,
	opts ...EuropePMCSearchOption,
) (*EuropePMCSearchResult, error) {
	if err := c.validate.Var(query, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	config := &europePMCSearchConfig{
		limit:      20,
		offset:     0,
		resultType: c.defaultResultType,
		format:     c.defaultFormat,
	}

	for _, opt := range opts {
		opt(config)
	}

	params := internal.EuropePMCSearchParams{
		Query:      query,
		ResultType: config.resultType,
		Format:     config.format,
		PageSize:   config.limit,
		CursorMark: "*",
	}

	apiResponse, err := c.europePMCService.SearchArticles(params)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeAPIError,
			Message: fmt.Sprintf("search failed: %s", err.Error()),
			Query:   query,
			Cause:   err,
		}
	}

	return convertFromInternalEuropePMCSearchResult(
		apiResponse,
		query,
		config.limit,
		config.offset,
	), nil
}

// FindSimilar finds articles similar to the given PMID using EuropePMC's related articles feature.
func (c *EuropePMCClient) FindSimilar(
	pmid string,
	opts ...EuropePMCSearchOption,
) (*EuropePMCSearchResult, error) {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	config := &europePMCSearchConfig{
		limit:      20,
		offset:     0,
		resultType: c.defaultResultType,
		format:     c.defaultFormat,
	}

	for _, opt := range opts {
		opt(config)
	}

	// Use EuropePMC's citation query to find related articles
	query := fmt.Sprintf("REF:\"%s\"", pmid)
	return c.Search(
		query,
		WithEuropePMCLimit(config.limit),
		WithEuropePMCOffset(config.offset),
	)
}

// HasPDF checks if a PDF is available for the given PMID in EuropePMC.
func (c *EuropePMCClient) HasPDF(pmid string) (bool, error) {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return false, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	article, err := c.GetArticle(pmid)
	if err != nil {
		return false, err
	}

	return article.HasPDF, nil
}

// GetPDFURLs retrieves available PDF URLs for the given PMID from EuropePMC.
func (c *EuropePMCClient) GetPDFURLs(
	pmid string,
) ([]EuropePMCFullTextURL, error) {
	if err := c.validate.Var(pmid, "required"); err != nil {
		return nil, &Error{
			Type:    ErrorTypeInvalidInput,
			Message: fmt.Sprintf("validation failed: %s", err.Error()),
		}
	}

	article, err := c.GetArticle(pmid)
	if err != nil {
		return nil, err
	}

	if !article.HasPDF {
		return nil, &Error{
			Type:    ErrorTypePDFNotAvailable,
			Message: "PDF not available for this article",
			PMID:    pmid,
		}
	}

	// Filter for PDF URLs
	var pdfURLs []EuropePMCFullTextURL
	for _, fullTextURL := range article.FullTextURLs {
		if fullTextURL.DocumentStyle == "pdf" {
			pdfURLs = append(pdfURLs, fullTextURL)
		}
	}

	if len(pdfURLs) == 0 {
		return nil, &Error{
			Type:    ErrorTypePDFNotFound,
			Message: "no PDF URLs found despite PDF availability flag",
			PMID:    pmid,
		}
	}

	return pdfURLs, nil
}

// convertFromInternalEuropePMCArticle converts internal API response to public type.
func convertFromInternalEuropePMCArticle(
	apiArticle *internal.EuropePMCAPIArticle,
) *EuropePMCArticle {
	article := &EuropePMCArticle{
		ID:           apiArticle.ID,
		Source:       apiArticle.Source,
		PMID:         apiArticle.PMID,
		PMCID:        apiArticle.PMCID,
		DOI:          apiArticle.DOI,
		Title:        apiArticle.Title,
		AuthorString: apiArticle.AuthorString,
		Abstract:     apiArticle.AbstractText,
		PubYear:      apiArticle.PubYear,
		PageInfo:     apiArticle.PageInfo,
		IsOpenAccess: apiArticle.IsOpenAccess == "Y",
		HasPDF:       apiArticle.HasPDF == "Y",
		License:      apiArticle.License,
		CitedByCount: apiArticle.CitedByCount,
		Language:     apiArticle.Language,
	}

	// Convert journal information
	article.Journal = convertJournalInfo(apiArticle.JournalInfo)

	// Convert authors
	article.Authors = convertAuthors(apiArticle.AuthorList.Author)

	// Convert publication types
	if len(apiArticle.PubTypeList.PubType) > 0 {
		article.PubTypes = apiArticle.PubTypeList.PubType
	}

	// Convert keywords
	if len(apiArticle.KeywordList.Keyword) > 0 {
		article.Keywords = apiArticle.KeywordList.Keyword
	}

	// Convert grants
	article.Grants = convertGrants(apiArticle.GrantsList.Grant)

	// Convert MeSH headings
	article.MeshHeadings = convertMeshHeadings(
		apiArticle.MeshHeadingList.MeshHeading,
	)

	// Convert chemicals
	article.Chemicals = convertChemicals(apiArticle.ChemicalList.Chemical)

	// Convert full text URLs
	article.FullTextURLs = convertFullTextURLs(
		apiArticle.FullTextURLList.FullTextURL,
	)

	// Parse dates
	convertArticleDates(article, apiArticle)

	return article
}

// convertJournalInfo converts internal journal info to public type.
func convertJournalInfo(journalInfo internal.EuropePMCAPIJournalInfo) EuropePMCJournal {
	return EuropePMCJournal{
		Title:               journalInfo.Journal.Title,
		MedlineAbbreviation: journalInfo.Journal.MedlineAbbreviation,
		ISOAbbreviation:     journalInfo.Journal.IsoAbbreviation,
		ISSN:                journalInfo.Journal.ISSN,
		ESSN:                journalInfo.Journal.ESSN,
		Volume:              journalInfo.Volume,
		Issue:               journalInfo.Issue,
		IssueID:             journalInfo.JournalIssueId,
		DateOfPublication:   journalInfo.DateOfPublication,
		MonthOfPublication:  journalInfo.MonthOfPublication,
		YearOfPublication:   journalInfo.YearOfPublication,
		NLMID:               journalInfo.Journal.NLMID,
	}
}

// convertArticleDates parses and sets date fields on the article.
func convertArticleDates(article *EuropePMCArticle, apiArticle *internal.EuropePMCAPIArticle) {
	article.CreationDate = parseDate(apiArticle.DateOfCreation)
	article.RevisionDate = parseDate(apiArticle.DateOfRevision)
	article.PublishDate = parseDate(apiArticle.FirstPublicationDate)
}

// convertFromInternalEuropePMCSearchResult converts internal API response to public search result type.
func convertFromInternalEuropePMCSearchResult(
	apiResponse *internal.EuropePMCAPIResponse,
	query string,
	limit, offset int,
) *EuropePMCSearchResult {
	result := &EuropePMCSearchResult{
		Query:  query,
		Total:  apiResponse.HitCount,
		Limit:  limit,
		Offset: offset,
	}

	for _, apiArticle := range apiResponse.ResultList.Result {
		article := convertFromInternalEuropePMCArticle(&apiArticle)
		result.Articles = append(result.Articles, article)
	}

	return result
}

// Helper functions for conversion

func convertAuthors(
	apiAuthors []internal.EuropePMCAPIAuthor,
) []EuropePMCAuthor {
	authors := make([]EuropePMCAuthor, 0, len(apiAuthors))
	for _, apiAuthor := range apiAuthors {
		author := EuropePMCAuthor{
			FullName:  apiAuthor.FullName,
			FirstName: apiAuthor.FirstName,
			LastName:  apiAuthor.LastName,
			Initials:  apiAuthor.Initials,
		}

		if apiAuthor.AuthorId.Type == "ORCID" {
			author.ORCID = apiAuthor.AuthorId.Value
		}

		for _, affiliation := range apiAuthor.AuthorAffiliationDetailsList.AuthorAffiliation {
			author.Affiliations = append(
				author.Affiliations,
				EuropePMCAuthorAffiliation{
					Affiliation: affiliation.Affiliation,
				},
			)
		}

		authors = append(authors, author)
	}
	return authors
}

func convertGrants(apiGrants []internal.EuropePMCAPIGrant) []EuropePMCGrant {
	grants := make([]EuropePMCGrant, 0, len(apiGrants))
	for _, apiGrant := range apiGrants {
		grant := EuropePMCGrant{
			GrantID: apiGrant.GrantId,
			Agency:  apiGrant.Agency,
			OrderIn: apiGrant.OrderIn,
		}
		grants = append(grants, grant)
	}
	return grants
}

func convertMeshHeadings(
	apiMeshHeadings []internal.EuropePMCAPIMeshHeading,
) []EuropePMCMeshHeading {
	meshHeadings := make([]EuropePMCMeshHeading, 0, len(apiMeshHeadings))
	for _, apiMesh := range apiMeshHeadings {
		mesh := EuropePMCMeshHeading{
			MajorTopic:     apiMesh.MajorTopicYN == "Y",
			DescriptorName: apiMesh.DescriptorName,
		}

		for _, apiQualifier := range apiMesh.MeshQualifierList.MeshQualifier {
			qualifier := EuropePMCMeshQualifier{
				Abbreviation:  apiQualifier.Abbreviation,
				QualifierName: apiQualifier.QualifierName,
				MajorTopic:    apiQualifier.MajorTopicYN == "Y",
			}
			mesh.MeshQualifiers = append(mesh.MeshQualifiers, qualifier)
		}

		meshHeadings = append(meshHeadings, mesh)
	}
	return meshHeadings
}

func convertChemicals(
	apiChemicals []internal.EuropePMCAPIChemical,
) []EuropePMCChemical {
	chemicals := make([]EuropePMCChemical, 0, len(apiChemicals))
	for _, apiChemical := range apiChemicals {
		chemical := EuropePMCChemical{
			Name:           apiChemical.Name,
			RegistryNumber: apiChemical.RegistryNumber,
		}
		chemicals = append(chemicals, chemical)
	}
	return chemicals
}

func convertFullTextURLs(
	apiURLs []internal.EuropePMCAPIFullTextURL,
) []EuropePMCFullTextURL {
	fullTextURLs := make([]EuropePMCFullTextURL, 0, len(apiURLs))
	for _, apiURL := range apiURLs {
		fullTextURL := EuropePMCFullTextURL{
			Availability:     apiURL.Availability,
			AvailabilityCode: apiURL.AvailabilityCode,
			DocumentStyle:    apiURL.DocumentStyle,
			Site:             apiURL.Site,
			URL:              apiURL.URL,
		}
		fullTextURLs = append(fullTextURLs, fullTextURL)
	}
	return fullTextURLs
}

func parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
		return &parsedDate
	}
	return nil
}
