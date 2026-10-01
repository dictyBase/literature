package literature

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dictybase/literature/internal"
	"github.com/stretchr/testify/require"
)

const (
	// testWorkPDFURL is the shared mock PDF URL for test fixtures.
	testWorkPDFURL = "https://example.org/pdf"
	// testJournalName is the shared mock journal name for test fixtures.
	testJournalName = "Test Journal"
	// testJournalISSN is the shared mock journal ISSN for test fixtures.
	testJournalISSN = "1234-5678"
	// testSecondRefID is the shared second referenced-work ID for fixtures.
	testSecondRefID = "W200"
)

// mockOpenAlexAPIWork builds a raw OpenAlex work response body.
//
//nolint:funlen // test fixture builder
func mockOpenAlexAPIWork(id string) *internal.OpenAlexAPIWork {
	return &internal.OpenAlexAPIWork{
		ID:              "https://openalex.org/" + id,
		DOI:             "https://doi.org/10.1000/test",
		OpenAlexIDs:     internal.OpenAlexAPIIdentifiers{PMID: "PMID:12345678"},
		DisplayName:     "Test Work",
		PublicationYear: 2024,
		PublicationDate: "2024-01-15",
		Type:            "journal-article",
		CitedByCount:    42,
		FWCI:            1.5,
		CitedByPercentileYear: internal.OpenAlexAPIPercentileYear{
			Min: 0.9,
			Max: 0.95,
		},
		ReferencedWorks: []string{"W100", "W200"},
		RelatedWorks:    []string{"W300"},
		IsRetracted:     false,
		OpenAccess: internal.OpenAlexAPIOpenAccess{
			IsOA:     true,
			OAStatus: "gold",
			OAURL:    testWorkPDFURL,
		},
		PrimaryLocation: internal.OpenAlexAPILocation{
			LandingPageURL: "https://example.org/landing",
			PDFURL:         testWorkPDFURL,
			License:        "cc-by",
			Source: internal.OpenAlexAPISourceInfo{
				ID:          "https://openalex.org/S137773608",
				DisplayName: testJournalName,
				ISSNL:       testJournalISSN,
				ISSN:        []string{"1234-5678", "8765-4321"},
			},
		},
		BestOALocation: &internal.OpenAlexAPILocation{
			LandingPageURL: "https://example.org/landing",
			PDFURL:         testWorkPDFURL,
			License:        "cc-by",
			Source: internal.OpenAlexAPISourceInfo{
				ID:          "https://openalex.org/S137773608",
				DisplayName: testJournalName,
			},
		},
		Authorships: []internal.OpenAlexAPIAuthorship{
			{
				AuthorPosition: "first",
				Author: internal.OpenAlexAPIAuthor{
					ID:          "https://openalex.org/A1",
					ORCID:       "https://orcid.org/0000-0002-1825-0097",
					DisplayName: "Jane Doe",
				},
				Institutions: []internal.OpenAlexAPIInstitution{
					{
						ID:          "https://openalex.org/I1",
						ROR:         "https://ror.org/01",
						DisplayName: "Test University",
						CountryCode: "US",
					},
				},
				RawAffiliationStrings: []string{"Dept. of Testing, Test University"},
			},
		},
		Topics: []internal.OpenAlexAPITopic{
			{
				ID:          "https://openalex.org/T1",
				DisplayName: "Topic One",
				Score:       0.8,
				Subfield: internal.OpenAlexAPITopicNode{
					ID:          "https://openalex.org/SF1",
					DisplayName: "Subfield One",
				},
				Field: internal.OpenAlexAPITopicNode{
					ID:          "https://openalex.org/F1",
					DisplayName: "Field One",
				},
				Domain: internal.OpenAlexAPITopicNode{
					ID:          "https://openalex.org/D1",
					DisplayName: "Domain One",
				},
			},
		},
		Keywords: []internal.OpenAlexAPIKeyword{
			{ID: "https://openalex.org/K1", DisplayName: "keyword", Score: 0.7},
		},
		Grants: []internal.OpenAlexAPIGrant{
			{
				FunderID:          "https://openalex.org/F4320306076",
				FunderDisplayName: "Test Funder",
				AwardID:           "R01-12345",
				AwardAmount:       500000,
				FundingYear:       2023,
			},
		},
		Biblio: internal.OpenAlexAPIBiblio{
			Volume: "10", Issue: "2", FirstPage: "1", LastPage: "12",
		},
	}
}

// mockOpenAlexWorksList builds a works list response body.
func mockOpenAlexWorksList(ids ...string) *internal.OpenAlexAPIWorksResponse {
	response := &internal.OpenAlexAPIWorksResponse{
		Meta: internal.OpenAlexAPIMeta{
			Count: len(ids), PerPage: 25, Cursor: "cursor-next",
		},
	}
	for _, id := range ids {
		response.Results = append(response.Results, *mockOpenAlexAPIWork(id))
	}
	return response
}

// createTestOpenAlexClient creates a test client with a mock server.
func createTestOpenAlexClient(
	t *testing.T,
	handler http.HandlerFunc,
	opts ...OpenAlexOption,
) *OpenAlexClient {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	opts = append(
		[]OpenAlexOption{
			WithOpenAlexHTTPClient(server.Client()),
			WithOpenAlexBaseURL(server.URL),
		},
		opts...,
	)

	client, err := NewOpenAlexClient(opts...)
	require.NoError(t, err)

	return client
}

// openAlexServeJSON writes a JSON response with 200.
func openAlexServeJSON(
	t *testing.T,
	w http.ResponseWriter,
	payload any,
) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(payload))
}

func TestGetWork_Success(t *testing.T) {
	t.Parallel()

	expected := mockOpenAlexAPIWork("W123")
	var gotPath string
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			openAlexServeJSON(t, w, expected)
		},
	)

	work, err := client.GetWork("W123")

	req := require.New(t)
	req.NoError(err)
	req.True(
		strings.HasSuffix(gotPath, "/works/W123"),
		"unexpected path: %s", gotPath,
	)
	req.Equal("W123", work.OpenAlexID)
	req.Equal("10.1000/test", work.DOI)
	req.Equal("12345678", work.PMID)
	req.Equal("Test Work", work.Title)
	req.Equal(2024, work.PublicationYear)
	req.Equal(42, work.CitedByCount)
	req.InEpsilon(1.5, work.FWCI, 0.0001)
	req.InEpsilon(0.9, work.PercentileMin, 0.0001)
	req.InEpsilon(0.95, work.PercentileMax, 0.0001)
	req.Equal([]string{"W100", "W200"}, work.ReferencedWorks)
	req.True(work.OpenAccess.IsOA)
	req.Equal("gold", work.OpenAccess.OAStatus)
	req.Equal("cc-by", work.PrimaryLocation.License)
	req.Equal("S137773608", work.PrimaryLocation.Source.OpenAlexID)
	req.Len(work.Authors, 1)
	req.Equal("0000-0002-1825-0097", work.Authors[0].Author.ORCID)
	req.Equal("01", work.Authors[0].Institutions[0].ROR)
	req.Len(work.Topics, 1)
	req.Equal("Domain One", work.Topics[0].Domain.DisplayName)
	req.Len(work.Keywords, 1)
	req.Len(work.Awards, 1)
	req.Equal("F4320306076", work.Awards[0].Funder.OpenAlexID)
	req.Equal("Test Funder", work.Awards[0].Funder.DisplayName)
	req.Equal("R01-12345", work.Awards[0].AwardID)
	req.Equal("10", work.Volume)
	req.NotNil(work.BestOALocation)
}

func TestGetWork_NotFound(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	_, err := client.GetWork("W404")

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, &Error{Type: ErrorTypeArticleNotFound})
	req.Contains(err.Error(), "work not found")
}

func TestGetWork_RateLimited(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		},
	)

	_, err := client.GetWork("W429")

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, &Error{Type: ErrorTypeRateLimit})
}

func TestGetWork_InvalidInput(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	_, err := client.GetWork("")

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, &Error{Type: ErrorTypeInvalidInput})
}

func TestGetWork_FullOpenAlexURL(t *testing.T) {
	t.Parallel()

	var gotPath string
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			openAlexServeJSON(t, w, mockOpenAlexAPIWork("W123"))
		},
	)

	_, err := client.GetWork("https://openalex.org/W123")

	req := require.New(t)
	req.NoError(err)
	req.True(
		strings.HasSuffix(gotPath, "/works/W123"),
		"unexpected path: %s", gotPath,
	)
}

// realisticWorkJSON mirrors the actual OpenAlex work response shape,
// including the array-valued source issn and flat grant funder fields.
const realisticWorkJSON = `{
	"id": "https://openalex.org/W2100837269",
	"doi": "https://doi.org/10.1038/nature12373",
	"ids": {"openalex": "https://openalex.org/W2100837269", "pmid": "PMID:23842501"},
	"display_name": "A realistic work",
	"publication_year": 2013,
	"publication_date": "2013-09-01",
	"type": "article",
	"cited_by_count": 117,
	"fwci": 3.28,
	"cited_by_percentile_year": {"min": 0.99, "max": 1.0},
	"referenced_works": ["https://openalex.org/W2103302949"],
	"open_access": {"is_oa": true, "oa_status": "hybrid"},
	"primary_location": {
		"landing_page_url": "https://example.org/landing",
		"pdf_url": "https://example.org/pdf",
		"license": "cc-by",
		"source": {
			"id": "https://openalex.org/S137773608",
			"display_name": "Test Journal",
			"issn_l": "0028-0836",
			"issn": ["0028-0836", "1476-4687"]
		}
	},
	"authorships": [{
		"author_position": "first",
		"author": {"id": "https://openalex.org/A1",
			"orcid": "https://orcid.org/0000-0002-1825-0097",
			"display_name": "Jane Doe"},
		"institutions": [{"id": "https://openalex.org/I1",
			"ror": "https://ror.org/01",
			"display_name": "Test University",
			"country_code": "US"}]
	}],
	"topics": [{
		"id": "https://openalex.org/T1",
		"display_name": "Topic One",
		"score": 0.8,
		"subfield": {"id": "https://openalex.org/SF1", "display_name": "Subfield One"},
		"field": {"id": "https://openalex.org/F1", "display_name": "Field One"},
		"domain": {"id": "https://openalex.org/D1", "display_name": "Domain One"}
	}],
	"grants": [{
		"funder": "F4320306076",
		"funder_display_name": "Test Funder",
		"award_id": "R01-12345",
		"award_amount": 500000,
		"funding_year": 2023
	}]
}`

func TestGetWork_DecodesRealisticResponse(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(realisticWorkJSON))
		},
	)

	work, err := client.GetWork("W2100837269")

	req := require.New(t)
	req.NoError(err)
	req.Equal("W2100837269", work.OpenAlexID)
	req.Equal("10.1038/nature12373", work.DOI)
	req.Equal("23842501", work.PMID)
	req.InEpsilon(3.28, work.FWCI, 0.0001)
	req.InEpsilon(0.99, work.PercentileMin, 0.0001)
	req.InEpsilon(1.0, work.PercentileMax, 0.0001)
	req.Equal("0028-0836", work.PrimaryLocation.Source.ISSN)
	req.Equal("hybrid", work.OpenAccess.OAStatus)
	req.Len(work.Authors, 1)
	req.Len(work.Topics, 1)
	req.Len(work.Awards, 1)
	req.Equal("F4320306076", work.Awards[0].Funder.OpenAlexID)
	req.Equal("Test Funder", work.Awards[0].Funder.DisplayName)
}

func TestGetWorks_ChunksLargeBatches(t *testing.T) {
	t.Parallel()

	// 101 identifiers -> three chunked requests (50/50/1).
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = fmt.Sprintf("W%03d", i)
	}

	var chunkSizes []int
	var gotPerPage []string
	response := mockOpenAlexWorksList()
	client := createTestOpenAlexClient(
		t,
		func(writer http.ResponseWriter, request *http.Request) {
			filter := request.URL.Query().Get("filter")
			gotPerPage = append(gotPerPage, request.URL.Query().Get("per-page"))
			chunkSizes = append(
				chunkSizes,
				strings.Count(filter, "|")+1,
			)
			openAlexServeJSON(t, writer, response)
		},
	)

	works, err := client.GetWorks(ids)

	req := require.New(t)
	req.NoError(err)
	req.Len(chunkSizes, 3)
	req.Equal([]int{50, 50, 1}, chunkSizes)
	req.Equal([]string{"50", "50", "1"}, gotPerPage)
	req.Len(works, 3*len(response.Results))
}

func TestGetWorkByPMID_NormalizesIdentifier(t *testing.T) {
	t.Parallel()

	var gotPath string
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			openAlexServeJSON(t, w, mockOpenAlexAPIWork("Wpmid"))
		},
	)

	_, err := client.GetWorkByPMID("12345678")

	req := require.New(t)
	req.NoError(err)
	req.True(
		strings.HasSuffix(gotPath, "/works/PMID:12345678"),
		"unexpected path: %s", gotPath,
	)
}

func TestGetWorkByDOI_PreservesIdentifier(t *testing.T) {
	t.Parallel()

	var gotPath string
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			openAlexServeJSON(t, w, mockOpenAlexAPIWork("Wdoi"))
		},
	)

	_, err := client.GetWorkByDOI("10.1000/test")

	req := require.New(t)
	req.NoError(err)
	req.True(
		strings.HasSuffix(gotPath, "/works/https://doi.org/10.1000/test"),
		"unexpected path: %s", gotPath,
	)
}

func TestGetWorks_BatchFilter(t *testing.T) {
	t.Parallel()

	var gotFilter string
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotFilter = r.URL.Query().Get("filter")
			openAlexServeJSON(
				t, w, mockOpenAlexWorksList("W1", "W2"),
			)
		},
	)

	works, err := client.GetWorks([]string{"W1", "W2"})

	req := require.New(t)
	req.NoError(err)
	req.Equal("ids.openalex:W1|W2", gotFilter)
	req.Len(works, 2)
	req.Equal("W1", works[0].OpenAlexID)
}

func TestGetWorks_EmptyList(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	_, err := client.GetWorks([]string{})

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, &Error{Type: ErrorTypeInvalidInput})
}

func TestGetWorksByPMIDs_BatchFilter(t *testing.T) {
	t.Parallel()

	var gotFilter string
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotFilter = r.URL.Query().Get("filter")
			openAlexServeJSON(
				t, w, mockOpenAlexWorksList("W1", "W2"),
			)
		},
	)

	works, err := client.GetWorksByPMIDs([]string{"1", "PMID:2"})

	req := require.New(t)
	req.NoError(err)
	req.Equal("pmid:PMID:1|PMID:2", gotFilter)
	req.Len(works, 2)
}

func TestGetReferencedWorks_ResolvesBatch(t *testing.T) {
	t.Parallel()

	referenced := mockOpenAlexAPIWork("W100")
	referenced.ReferencedWorks = []string{
		"https://openalex.org/W100",
		testSecondRefID,
	}
	var gotFilter string
	client := createTestOpenAlexClient(
		t,
		func(writer http.ResponseWriter, request *http.Request) {
			if strings.HasSuffix(request.URL.Path, "/works/W100") {
				openAlexServeJSON(t, writer, referenced)
				return
			}
			gotFilter = request.URL.Query().Get("filter")
			openAlexServeJSON(t, writer, mockOpenAlexWorksList("W100", "W200"))
		},
	)

	works, err := client.GetReferencedWorks("W100")

	req := require.New(t)
	req.NoError(err)
	req.Equal("ids.openalex:W100|W200", gotFilter)
	req.Len(works, 2)
}

func TestGetReferencedWorks_EmptyReferences(t *testing.T) {
	t.Parallel()

	work := mockOpenAlexAPIWork("Wnone")
	work.ReferencedWorks = nil
	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			openAlexServeJSON(t, w, work)
		},
	)

	works, err := client.GetReferencedWorks("Wnone")

	req := require.New(t)
	req.NoError(err)
	req.Empty(works)
}

func TestGetCitingWorks_CitesFilter(t *testing.T) {
	t.Parallel()

	var gotFilter, gotPerPage, gotSort string
	client := createTestOpenAlexClient(
		t,
		func(writer http.ResponseWriter, request *http.Request) {
			query := request.URL.Query()
			gotFilter = query.Get("filter")
			gotPerPage = query.Get("per-page")
			gotSort = query.Get("sort")
			openAlexServeJSON(
				t, writer, mockOpenAlexWorksList("W5", "W6"),
			)
		},
	)

	result, err := client.GetCitingWorks(
		"W123",
		WithOpenAlexPerPage(10),
		WithOpenAlexSort("cited_by_count:desc"),
	)

	req := require.New(t)
	req.NoError(err)
	req.Equal("cites:W123", gotFilter)
	req.Equal("10", gotPerPage)
	req.Equal("cited_by_count:desc", gotSort)
	req.Len(result.Results, 2)
	req.Equal(2, result.TotalCount)
	req.Equal("cursor-next", result.Cursor)
}

func TestGetCitationMetrics_Success(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			openAlexServeJSON(t, w, mockOpenAlexAPIWork("Wmetrics"))
		},
	)

	metrics, err := client.GetCitationMetrics("Wmetrics")

	req := require.New(t)
	req.NoError(err)
	req.Equal(42, metrics.CitedByCount)
	req.InEpsilon(1.5, metrics.FWCI, 0.0001)
	req.InEpsilon(0.9, metrics.PercentileMin, 0.0001)
	req.InEpsilon(0.95, metrics.PercentileMax, 0.0001)
}

func TestGetCitationMetrics_NotFound(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	_, err := client.GetCitationMetrics("Wmissing")

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, &Error{Type: ErrorTypeArticleNotFound})
}

func TestOpenAlexClient_APIKeyNotLeakedInError(t *testing.T) {
	t.Parallel()

	client := createTestOpenAlexClient(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		WithOpenAlexAPIKey("identity-redaction-fixture"),
	)

	_, err := client.GetWork("Wfail")

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), "identity-redaction-fixture")
}
