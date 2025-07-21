package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dictybase/literature"
)

func main() {
	// Create a new EuropePMC client
	client, err := literature.NewEuropePMCClient(
		literature.WithEuropePMCTimeout(30*time.Second),
		literature.WithEuropePMCUserAgent("test-client/1.0"),
	)
	if err != nil {
		log.Fatalf("Failed to create EuropePMC client: %v", err)
	}

	// Test GetArticle with the PMID from our sample data
	pmid := "40602797"
	fmt.Printf("Fetching article with PMID: %s\n", pmid)

	article, err := client.GetArticle(pmid)
	if err != nil {
		log.Fatalf("Failed to get article: %v", err)
	}

	// Print some basic information
	fmt.Printf("Article ID: %s\n", article.ID)
	fmt.Printf("PMID: %s\n", article.PMID)
	fmt.Printf("Title: %s\n", article.Title)
	fmt.Printf("Authors: %s\n", article.AuthorString)
	fmt.Printf("Journal: %s\n", article.Journal.Title)
	fmt.Printf("Year: %s\n", article.PubYear)
	fmt.Printf("DOI: %s\n", article.DOI)
	fmt.Printf("Abstract: %.100s...\n", article.Abstract)
	fmt.Printf("Has PDF: %t\n", article.HasPDF)
	fmt.Printf("Is Open Access: %t\n", article.IsOpenAccess)
	fmt.Printf("Number of authors: %d\n", len(article.Authors))
	fmt.Printf("Keywords: %v\n", article.Keywords)

	if len(article.Authors) > 0 {
		fmt.Printf("First author: %s %s\n", article.Authors[0].FirstName, article.Authors[0].LastName)
		if article.Authors[0].ORCID != "" {
			fmt.Printf("First author ORCID: %s\n", article.Authors[0].ORCID)
		}
	}

	fmt.Printf("\nFull text URLs:\n")
	for _, url := range article.FullTextURLs {
		fmt.Printf("  %s (%s): %s\n", url.DocumentStyle, url.Site, url.URL)
	}

	fmt.Println("\nTest completed successfully!")
}
