package literature

import (
	"testing"
	"time"

	"github.com/dictybase/literature/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// internalArticleFixture builds the internal PubMed shape the adapter
// reads, with the given year and month on the journal issue.
func internalArticleFixture(year, month string) *internal.PubMedArticle {
	article := &internal.PubMedArticle{}
	article.MedlineCitation.PMID = "4893433"
	article.MedlineCitation.Article.ArticleTitle = "A title"
	article.MedlineCitation.Article.Journal.Title = "A journal"
	article.MedlineCitation.Article.Journal.JournalIssue.PubDate.Year = year
	article.MedlineCitation.Article.Journal.JournalIssue.PubDate.Month = month
	article.MedlineCitation.Article.AuthorList.Authors = []internal.Author{
		{LastName: "Huber", ForeName: "Robert J"},
	}
	article.PubmedData.ArticleIdList.ArticleIDs = []internal.ArticleID{
		{IDType: "doi", Value: "10.1016/j.bbamcr.2018.07.017"},
	}

	return article
}

func TestConvertFromInternalArticlePublishDateWithAbbreviatedMonth(
	t *testing.T,
) {
	t.Parallel()

	article := convertFromInternalArticle(internalArticleFixture("2018", "Jul"))

	assert.Equal(
		t,
		time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
		article.PublishDate,
	)
}

func TestConvertFromInternalArticlePublishDateWithNumericMonth(t *testing.T) {
	t.Parallel()

	article := convertFromInternalArticle(internalArticleFixture("2018", "7"))

	assert.Equal(
		t,
		time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
		article.PublishDate,
	)
}

func TestConvertFromInternalArticlePublishDateWithoutMonth(t *testing.T) {
	t.Parallel()

	article := convertFromInternalArticle(internalArticleFixture("2018", ""))

	assert.Equal(
		t,
		time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
		article.PublishDate,
	)
}

func TestConvertFromInternalArticlePublishDateWithoutYear(t *testing.T) {
	t.Parallel()

	article := convertFromInternalArticle(internalArticleFixture("", "Jul"))

	assert.True(
		t,
		article.PublishDate.IsZero(),
		"a missing year must leave the publish date unset",
	)
}

func TestConvertFromInternalArticlePublishDateWithInvalidYear(t *testing.T) {
	t.Parallel()

	article := convertFromInternalArticle(
		internalArticleFixture("unknown", "Jul"),
	)

	assert.True(
		t,
		article.PublishDate.IsZero(),
		"an unparsable year must leave the publish date unset",
	)
}

func TestConvertFromInternalArticleFields(t *testing.T) {
	t.Parallel()

	article := convertFromInternalArticle(internalArticleFixture("2018", "Jul"))
	require.Equal(t, "4893433", article.PMID)
	require.Equal(t, "A title", article.Title)
	require.Equal(t, "A journal", article.Journal)
	require.Equal(t, "10.1016/j.bbamcr.2018.07.017", article.DOI)
	require.Len(t, article.Authors, 1)
	assert.Equal(t, "Huber", article.Authors[0].LastName)
	assert.Equal(t, "Robert J", article.Authors[0].FirstName)
	assert.Equal(t, "Robert J Huber", article.Authors[0].FullName)
}
