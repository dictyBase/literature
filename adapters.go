package literature

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dictybase/literature/internal"
)

// Conversion functions to transform internal types to public API types

// convertFromInternalArticle converts internal.PubMedArticle to public Article
func convertFromInternalArticle(internal *internal.PubMedArticle) *Article {
	article := &Article{
		PMID:     internal.GetPMID(),
		Title:    internal.GetTitle(),
		Abstract: internal.GetAbstract(),
		Journal:  internal.GetJournalTitle(),
	}

	// Get DOI if available
	if doi, exists := internal.GetDOI(); exists {
		article.DOI = doi
	}

	// Convert authors
	for _, author := range internal.GetAuthors() {
		article.Authors = append(article.Authors, Author{
			FirstName: author.ForeName,
			LastName:  author.LastName,
			FullName:  strings.TrimSpace(fmt.Sprintf("%s %s", author.ForeName, author.LastName)),
		})
	}

	return article
}

// convertFromInternalSearchResult converts internal.ESearchResult to public SearchResult
func convertFromInternalSearchResult(internal *internal.ESearchResult, query string, limit, offset int) *SearchResult {
	total := 0
	if internal.Count != "" {
		if count, err := strconv.Atoi(internal.Count); err == nil {
			total = count
		}
	}

	return &SearchResult{
		Query:    query,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		Articles: []*Article{}, // Would be populated by fetching each PMID if needed
	}
}
