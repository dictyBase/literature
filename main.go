package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v2" // Make sure you have this package: go get github.com/urfave/cli/v2
)

func searchAction(c *cli.Context) error {
	searchTerm := c.String("query")
	slog.Info("Searching PubMed", "query", searchTerm)

	esearchResult, err := searchPubMed(searchTerm)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if len(esearchResult.IDList.IDs) == 0 {
		slog.Info("No PubMed IDs found for the query.")
		return nil
	}

	slog.Info(
		"Found articles",
		"count", esearchResult.Count,
		"retrieving", len(esearchResult.IDList.IDs),
	)

	articleSet, err := fetchPubMedDetails(
		esearchResult.WebEnv,
		esearchResult.QueryKey,
	)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	var results strings.Builder
	results.WriteString("--- Retrieved Articles (PMID and Title) ---\n")
	for i, article := range articleSet.PubMedArticles {
		results.WriteString(
			fmt.Sprintf(
				"%d. PMID: %s\n",
				i+1,
				article.MedlineCitation.PMID,
			),
		)
		results.WriteString(fmt.Sprintf(
			"   Title: %s\n",
			article.MedlineCitation.Article.ArticleTitle,
		))
		results.WriteString(
			"--------------------------------------------\n",
		)
	}

	fmt.Print(results.String())

	return nil
}

func getArticleAction(c *cli.Context) error {
	if c.NArg() == 0 {
		return cli.Exit("PubMed ID is required.", 1)
	}
	pmid := c.Args().First()
	slog.Info("Fetching PubMed article", "pmid", pmid)

	articleSet, err := fetchPubMedArticle(pmid)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if len(articleSet.PubMedArticles) == 0 {
		slog.Info("No PubMed article found for the given ID.")
		return nil
	}

	article := articleSet.PubMedArticles[0]
	fmt.Printf("Title: %s\n", article.MedlineCitation.Article.ArticleTitle)

	pubDate := fmt.Sprintf(
		"%s %s",
		article.MedlineCitation.Article.Journal.JournalIssue.PubDate.Month,
		article.MedlineCitation.Article.Journal.JournalIssue.PubDate.Year,
	)
	fmt.Printf("Publication Date: %s\n", pubDate)

	fmt.Printf("Journal: %s\n", article.MedlineCitation.Article.Journal.Title)

	authors := Map(
		article.MedlineCitation.Article.AuthorList.Authors,
		formatAuthor,
	)
	fmt.Printf("Authors: %s\n", strings.Join(authors, ", "))

	fmt.Printf(
		"Pages: %s\n",
		article.MedlineCitation.Article.Pagination.MedlinePgn,
	)

	doiArticleID, found := Find(
		article.PubmedData.ArticleIdList.ArticleIds,
		isDOI,
	)

	if found {
		fmt.Printf("DOI: %s\n", doiArticleID.Value)
	}

	fmt.Printf(
		"Abstract: %s\n",
		article.MedlineCitation.Article.Abstract.AbstractText,
	)

	return nil
}

func main() {
	app := &cli.App{
		Name:  "pubmed",
		Usage: "A tool to interact with PubMed.",
		Commands: []*cli.Command{
			{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "Search PubMed for articles and list their IDs and titles.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "query",
						Aliases:  []string{"q"},
						Usage:    "Search term for PubMed (e.g., 'CRISPR gene editing')",
						Required: true,
					},
				},
				Action: searchAction,
			},
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "Get article details for a given PubMed ID.",
				ArgsUsage: "<PubMed ID>",
				Action:    getArticleAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("application failed to run", "error", err)
		os.Exit(1)
	}
}
