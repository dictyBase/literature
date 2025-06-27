package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	t.Run("TestSearchPubMedWithLimit_Success", func(t *testing.T) {
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
