package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOpenAlexWork builds a minimal raw work object for tests.
func mockOpenAlexWork(id string) *OpenAlexAPIWork {
	return &OpenAlexAPIWork{
		ID:              fmt.Sprintf("https://openalex.org/%s", id),
		DOI:             "https://doi.org/10.1000/test",
		OpenAlexIDs:     OpenAlexAPIIdentifiers{PMID: "PMID:12345678"},
		DisplayName:     "Test Work",
		PublicationYear: 2024,
		CitedByCount:    42,
		FWCI:            1.5,
		CitedByPercentileYear: OpenAlexAPIPercentileYear{
			Min: 0.9,
			Max: 0.95,
		},
		ReferencedWorks: []string{"W1", "W2"},
	}
}

// newOpenAlexTestService creates a service pointed at the mock server.
func newOpenAlexTestService(
	t *testing.T,
	handler http.HandlerFunc,
	opts ...OpenAlexServiceOption,
) *OpenAlexService {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	opts = append(
		[]OpenAlexServiceOption{
			WithOpenAlexHTTPClient(server.Client()),
			WithOpenAlexBaseURL(server.URL),
		},
		opts...,
	)

	return NewOpenAlexService(opts...)
}

func TestFetchWork_Success(t *testing.T) {
	t.Parallel()

	expected := mockOpenAlexWork("W123")
	var gotPath string
	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			assert.NoError(t, json.NewEncoder(w).Encode(expected))
		},
	)

	work, err := service.FetchWork("W123")

	req := require.New(t)
	req.NoError(err)
	req.True(
		strings.HasSuffix(gotPath, "/works/W123"),
		"unexpected path: %s", gotPath,
	)
	req.Equal(expected.ID, work.ID)
	req.Equal(expected.CitedByCount, work.CitedByCount)
	req.Equal(expected.ReferencedWorks, work.ReferencedWorks)
}

func TestFetchWork_NotFound(t *testing.T) {
	t.Parallel()

	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	_, err := service.FetchWork("W404")

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, ErrOpenAlexNotFound)
}

func TestFetchWork_RateLimited(t *testing.T) {
	t.Parallel()

	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		},
	)

	_, err := service.FetchWork("W429")

	req := require.New(t)
	req.Error(err)
	req.ErrorIs(err, ErrOpenAlexRateLimited)
}

func TestFetchWork_ServerError(t *testing.T) {
	t.Parallel()

	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	)

	_, err := service.FetchWork("W500")

	req := require.New(t)
	req.Error(err)
	req.Contains(err.Error(), "500")
	req.NotErrorIs(err, ErrOpenAlexNotFound)
}

func TestFetchWork_Timeout(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	handler := func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
	}
	service := newOpenAlexTestService(
		t,
		handler,
		WithOpenAlexHTTPClient(&http.Client{
			Timeout: 50 * time.Millisecond,
		}),
	)
	t.Cleanup(func() { close(release) })

	_, err := service.FetchWork("Wslow")

	req := require.New(t)
	req.Error(err)
	var urlError interface{ Error() string } = err
	req.NotEmpty(urlError.Error())
}

func TestFetchWork_QueryCredentials(t *testing.T) {
	t.Parallel()

	var gotQuery string
	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			assert.NoError(t, json.NewEncoder(w).Encode(mockOpenAlexWork("W1")))
		},
		WithOpenAlexEmail("polite@example.com"),
		WithOpenAlexAPIKey("super-secret-key"),
	)

	_, err := service.FetchWork("W1")

	req := require.New(t)
	req.NoError(err)
	req.Contains(gotQuery, "mailto=polite%40example.com")
	req.Contains(gotQuery, "api_key=super-secret-key")
}

func TestFetchWork_DOIPath(t *testing.T) {
	t.Parallel()

	var gotPath string
	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			assert.NoError(t, json.NewEncoder(w).Encode(mockOpenAlexWork("Wdoi")))
		},
	)

	_, err := service.FetchWork("10.1000/test")

	req := require.New(t)
	req.NoError(err)
	req.True(
		strings.HasSuffix(gotPath, "/works/https://doi.org/10.1000/test"),
		"unexpected path: %s", gotPath,
	)
}

func TestFetchWork_EmptyIdentifier(t *testing.T) {
	t.Parallel()

	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	_, err := service.FetchWork("  ")

	req := require.New(t)
	req.Error(err)
}

func TestSearchWorks_Success(t *testing.T) {
	t.Parallel()

	response := &OpenAlexAPIWorksResponse{
		Meta:    OpenAlexAPIMeta{Count: 2, PerPage: 25, Cursor: "*"},
		Results: []OpenAlexAPIWork{*mockOpenAlexWork("W1"), *mockOpenAlexWork("W2")},
	}

	var gotFilter, gotPerPage string
	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			gotFilter = r.URL.Query().Get("filter")
			gotPerPage = r.URL.Query().Get("per-page")
			w.Header().Set("Content-Type", "application/json")
			assert.NoError(t, json.NewEncoder(w).Encode(response))
		},
	)

	params := OpenAlexWorksParams{
		Filter:  "ids.openalex:W1|W2",
		PerPage: 25,
	}

	result, err := service.SearchWorks(params)

	req := require.New(t)
	req.NoError(err)
	req.Equal("ids.openalex:W1|W2", gotFilter)
	req.Equal("25", gotPerPage)
	req.Len(result.Results, 2)
	req.Equal(2, result.Meta.Count)
}

func TestSearchWorks_DecodeFailure(t *testing.T) {
	t.Parallel()

	service := newOpenAlexTestService(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not json"))
		},
	)

	_, err := service.SearchWorks(OpenAlexWorksParams{Filter: "cites:W1"})

	req := require.New(t)
	req.Error(err)
	req.Contains(err.Error(), "decode")
}

func TestOpenAlexService_ErrorDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NotFoundHandler())
	client := &http.Client{
		Transport: &http.Transport{},
	}
	server.Close()

	service := NewOpenAlexService(
		WithOpenAlexHTTPClient(client),
		WithOpenAlexBaseURL("http://127.0.0.1:1"),
		WithOpenAlexAPIKey("identity-redaction-fixture"),
	)

	// Force a transport error against an unreachable server.
	_, err := service.FetchWork("W1")

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), "identity-redaction-fixture")
}

func TestOpenAlexService_InvalidRequestURLDoesNotLeakAPIKey(t *testing.T) {
	t.Parallel()

	service := NewOpenAlexService(
		WithOpenAlexBaseURL("http://%zz"),
		WithOpenAlexAPIKey("identity-redaction-fixture"),
	)

	_, err := service.FetchWork("W1")

	req := require.New(t)
	req.Error(err)
	req.NotContains(err.Error(), "identity-redaction-fixture")
}

func TestOpenAlexService_NilHTTPOptionIsIgnored(t *testing.T) {
	t.Parallel()

	service := NewOpenAlexService(WithOpenAlexHTTPClient(nil))

	req := require.New(t)
	req.NotNil(service.httpClient)
}

func TestOpenAlexService_Defaults(t *testing.T) {
	t.Parallel()

	service := NewOpenAlexService()

	req := require.New(t)
	req.Equal("https://api.openalex.org", service.baseURL)
	req.Equal("openalex-go-client/1.0", service.userAgent)
	req.NotNil(service.httpClient)
	req.NotZero(service.httpClient.Timeout)
}
