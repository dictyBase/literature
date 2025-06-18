package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
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
	XMLName        xml.Name        `xml:"PubMedArticleSet"`
	PubMedArticles []PubMedArticle `xml:"PubMedArticle"`
}

// PubMedArticle represents a single PubMed article.
type PubMedArticle struct {
	XMLName         xml.Name `xml:"PubMedArticle"`
	MedlineCitation struct {
		Article struct {
			ArticleTitle string `xml:"ArticleTitle"`
		} `xml:"Article"`
		PMID string `xml:"PMID"`
	} `xml:"MedlineCitation"`
}

func searchPubMed(query string) (*ESearchResult, error) {
	esearchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=%s&retmax=10&retmode=xml",
		url.QueryEscape(query),
	)

	// #nosec G107
	resp, err := http.Get(esearchURL)
	if err != nil {
		return nil, fmt.Errorf("error making esearch request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading esearch response body: %w", err)
	}

	var esearchResult ESearchResult
	if err := xml.Unmarshal(body, &esearchResult); err != nil {
		return nil, fmt.Errorf(
			"error unmarshaling esearch XML: %w\nXML:\n%s",
			err,
			string(body),
		)
	}

	return &esearchResult, nil
}

func fetchPubMedDetails(ids []string) (*PubMedArticleSet, error) {
	idString := strings.Join(ids, ",")
	efetchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&id=%s&retmode=xml",
		idString,
	)

	// #nosec G107
	resp, err := http.Get(efetchURL)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading efetch response body: %w", err)
	}

	var articleSet PubMedArticleSet
	if err := xml.Unmarshal(body, &articleSet); err != nil {
		return nil, fmt.Errorf(
			"error unmarshaling efetch XML: %w\nXML:\n%s",
			err,
			string(body),
		)
	}

	return &articleSet, nil
}

func main() {
	app := &cli.App{
		Name:  "pubmed-search",
		Usage: "Search PubMed for articles and list their IDs and titles.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "query",
				Aliases:  []string{"q"},
				Usage:    "Search term for PubMed (e.g., 'CRISPR gene editing')",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			searchTerm := c.String("query")
			if searchTerm == "" {
				return cli.Exit(
					"Error: A search query is required. Use --query or -q.",
					1,
				)
			}

			fmt.Printf("Searching PubMed for: '%s'\n", searchTerm)

			esearchResult, err := searchPubMed(searchTerm)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			if len(esearchResult.IDList.IDs) == 0 {
				fmt.Println("No PubMed IDs found for the query.")
				return nil
			}

			fmt.Printf(
				"Found %s articles. Retrieving titles for the first %d...\n\n",
				esearchResult.Count,
				len(esearchResult.IDList.IDs),
			)

			articleSet, err := fetchPubMedDetails(esearchResult.IDList.IDs)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			fmt.Println("--- Retrieved Articles (PMID and Title) ---")
			for i, article := range articleSet.PubMedArticles {
				fmt.Printf("%d. PMID: %s\n", i+1, article.MedlineCitation.PMID)
				fmt.Printf(
					"   Title: %s\n",
					article.MedlineCitation.Article.ArticleTitle,
				)
				fmt.Println("--------------------------------------------")
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
