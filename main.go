package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/urfave/cli/v2" // Make sure you have this package: go get github.com/urfave/cli/v2
)

// ESearchResult represents the top-level XML structure for esearch.
type ESearchResult struct {
	XMLName  xml.Name `xml:"eSearchResult"`
	Count    string   `xml:"Count"`
	RetMax   string   `xml:"RetMax"`
	RetStart string   `xml:"RetStart"`
	QueryKey string   `xml:"QueryKey"`
	WebEnv   string   `xml:"WebEnv"`
	IDList   struct {
		IDs []string `xml:"Id"`
	} `xml:"IdList"`
}

// PubMedArticleSet represents the top-level XML structure for efetch.
type PubMedArticleSet struct {
	XMLName        xml.Name        `xml:"PubmedArticleSet"`
	PubMedArticles []PubMedArticle `xml:"PubmedArticle"`
}

// PubMedArticle represents a single PubMed article.
type PubMedArticle struct {
	XMLName         xml.Name `xml:"PubmedArticle"`
	MedlineCitation struct {
		Article struct {
			ArticleTitle string `xml:"ArticleTitle"`
			Abstract     struct {
				AbstractText string `xml:"AbstractText"`
			} `xml:"Abstract"`
		} `xml:"Article"`
		PMID string `xml:"PMID"`
	} `xml:"MedlineCitation"`
}

func searchPubMed(query string) (*ESearchResult, error) {
	esearchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=%s&retmax=10&retmode=xml&usehistory=y",
		url.QueryEscape(query),
	)

	// #nosec G107
	resp, err := http.Get(esearchURL)
	if err != nil {
		return nil, fmt.Errorf("error making esearch request: %w", err)
	}
	defer resp.Body.Close()

	esearchResult := &ESearchResult{}
	if err := xml.NewDecoder(resp.Body).Decode(esearchResult); err != nil {
		return nil, fmt.Errorf("error unmarshaling esearch XML: %w", err)
	}

	return esearchResult, nil
}

func fetchPubMedDetails(webEnv, queryKey string) (*PubMedArticleSet, error) {
	efetchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&retmode=xml&WebEnv=%s&query_key=%s&retmax=10",
		webEnv,
		queryKey,
	)

	// #nosec G107
	resp, err := http.Get(efetchURL)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", err)
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling efetch XML: %w", err)
	}

	return articleSet, nil
}

func fetchPubMedArticle(pmid string) (*PubMedArticleSet, error) {
	efetchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&retmode=xml&id=%s",
		pmid,
	)

	// #nosec G107
	resp, err := http.Get(efetchURL)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", err)
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling efetch XML: %w", err)
	}

	return articleSet, nil
}

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
