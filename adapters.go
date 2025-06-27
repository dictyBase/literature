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
			FullName: strings.TrimSpace(
				fmt.Sprintf("%s %s", author.ForeName, author.LastName),
			),
		})
	}

	return article
}

// convertFromInternalSearchResultWithArticles converts internal search results
// with detailed article information.
func convertFromInternalSearchResultWithArticles(
	searchResult *internal.ESearchResult,
	articleSet *internal.PubMedArticleSet,
	query string,
	limit, offset int,
) *SearchResult {
	total := 0
	if searchResult.Count != "" {
		if count, err := strconv.Atoi(searchResult.Count); err == nil {
			total = count
		}
	}

	articles := make([]*Article, 0, len(articleSet.PubMedArticles))
	for _, internalArticle := range articleSet.PubMedArticles {
		articles = append(
			articles,
			convertFromInternalArticle(&internalArticle),
		)
	}

	return &SearchResult{
		Query:    query,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		Articles: articles,
	}
}
