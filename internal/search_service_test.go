package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSearchService(
	handler http.HandlerFunc,
) (*SearchService, *httptest.Server) {
	server := httptest.NewServer(handler)
	service := &SearchService{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}
	return service, server
}

func TestSearchPubMedWithLimit(t *testing.T) {
	t.Parallel()
	t.Run("TestSearchPubMedWithLimit_Success", func(t *testing.T) {
		t.Parallel()
		assrt := assert.New(t)
		const query = "biology"
		const limit = 5
		const offset = 2

		xmlResponse := `
		<eSearchResult>
			<Count>100</Count>
			<RetMax>5</RetMax>
			<RetStart>2</RetStart>
			<IdList>
				<Id>1</Id>
				<Id>2</Id>
				<Id>3</Id>
				<Id>4</Id>
				<Id>5</Id>
			</IdList>
		</eSearchResult>`

		service, server := newTestSearchService(
			http.HandlerFunc(
				func(writer http.ResponseWriter, request *http.Request) {
					assrt.Equal("GET", request.Method)
					assrt.Contains(
						request.URL.String(),
						fmt.Sprintf("term=%s", query),
					)
					assrt.Contains(
						request.URL.String(),
						fmt.Sprintf("retmax=%d", limit),
					)
					assrt.Contains(
						request.URL.String(),
						fmt.Sprintf("retstart=%d", offset),
					)
					writer.Header().Set("Content-Type", "application/xml")
					fmt.Fprint(writer, xmlResponse)
				},
			),
		)
		defer server.Close()

		req := require.New(t)
		result, err := service.SearchPubMed(query, limit, offset)
		req.NoError(err)
		req.NotNil(result)
		req.Equal("100", result.Count)
		req.Equal("5", result.RetMax)
		req.Equal("2", result.RetStart)
		req.Len(result.IDList.IDs, 5)
	})
}

func TestSearchPubMed_ConcurrentRequestsUsePerCallParameters(t *testing.T) {
	t.Parallel()
	const requestCount = 32
	type searchRequest struct {
		term     string
		retmax   string
		retstart string
	}
	expected := make(map[string]searchRequest, requestCount)
	for index := range requestCount {
		term := fmt.Sprintf("term-%d", index)
		expected[term] = searchRequest{
			term:     term,
			retmax:   fmt.Sprintf("%d", index+1),
			retstart: fmt.Sprintf("%d", index*10),
		}
	}
	captured := make(chan searchRequest, requestCount)
	service, server := newTestSearchService(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			query := request.URL.Query()
			captured <- searchRequest{
				term:     query.Get("term"),
				retmax:   query.Get("retmax"),
				retstart: query.Get("retstart"),
			}
			_, _ = fmt.Fprint(writer, "<eSearchResult></eSearchResult>")
		},
	))
	defer server.Close()

	start := make(chan struct{})
	requestErrors := make(chan error, requestCount)
	var waitGroup sync.WaitGroup
	for index := range requestCount {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			<-start
			term := fmt.Sprintf("term-%d", index)
			_, err := service.SearchPubMed(term, index+1, index*10)
			requestErrors <- err
		}(index)
	}
	close(start)
	waitGroup.Wait()
	close(requestErrors)

	req := require.New(t)
	for requestErr := range requestErrors {
		req.NoError(requestErr)
	}
	for range requestCount {
		actual := <-captured
		want, ok := expected[actual.term]
		req.True(ok, "unexpected search term %q", actual.term)
		req.Equal(want, actual)
	}
}
