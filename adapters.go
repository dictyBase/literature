package literature

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dictybase/literature/internal"
)

// monthNames maps PubMed's three-letter month abbreviations to months.
var monthNames = map[string]time.Month{
	"jan": time.January,
	"feb": time.February,
	"mar": time.March,
	"apr": time.April,
	"may": time.May,
	"jun": time.June,
	"jul": time.July,
	"aug": time.August,
	"sep": time.September,
	"oct": time.October,
	"nov": time.November,
	"dec": time.December,
}

// pubDate builds the publication date from PubMed's year and month
// parts, returning the zero time when the year is missing or
// unparsable.
func pubDate(year, month string) time.Time {
	parsedYear, err := strconv.Atoi(strings.TrimSpace(year))
	if err != nil {
		return time.Time{}
	}

	return time.Date(
		parsedYear,
		pubMonth(month),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

// pubMonth resolves a PubMed month, which is either a number or a
// three-letter abbreviation, defaulting to January.
func pubMonth(month string) time.Month {
	trimmed := strings.ToLower(strings.TrimSpace(month))

	parsed, err := strconv.Atoi(trimmed)
	if err == nil && parsed >= 1 && parsed <= 12 {
		return time.Month(parsed)
	}

	if len(trimmed) >= 3 {
		if resolved, found := monthNames[trimmed[:3]]; found {
			return resolved
		}
	}

	return time.January
}

// convertFromInternalArticle converts internal.PubMedArticle to public Article.
func convertFromInternalArticle(internal *internal.PubMedArticle) *Article {
	article := &Article{
		PMID:        internal.GetPMID(),
		Title:       internal.GetTitle(),
		Abstract:    internal.GetAbstract(),
		Journal:     internal.GetJournalTitle(),
		PublishDate: pubDate(internal.GetPubYear(), internal.GetPubMonth()),
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
