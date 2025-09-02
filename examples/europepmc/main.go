package main

import (
	"fmt"
	"log"

	"github.com/dictybase/literature"
)

func main() {
	// Create a new EuropePMC client
	client, err := literature.NewEuropePMCClient()
	if err != nil {
		log.Fatalf("Failed to create EuropePMC client: %v", err)
	}

	// Example PMID and DOI
	const examplePMID = "23842501"
	const exampleDOI = "10.1038/nature12373"

	// Fetch article by PMID
	fmt.Println("Fetching article by PMID:", examplePMID)
	article, err := client.GetArticle(examplePMID)
	if err != nil {
		log.Fatalf("Failed to fetch article by PMID: %v", err)
	}
	fmt.Printf("Title: %s\n", article.Title)
	fmt.Printf("DOI: %s\n", article.DOI)
	fmt.Printf("Authors: %s\n", article.AuthorString)
	fmt.Println()

	// Fetch article by DOI
	fmt.Println("Fetching article by DOI:", exampleDOI)
	articleByDOI, err := client.GetArticleByDOI(exampleDOI)
	if err != nil {
		log.Fatalf("Failed to fetch article by DOI: %v", err)
	}
	fmt.Printf("Title: %s\n", articleByDOI.Title)
	fmt.Printf("PMID: %s\n", articleByDOI.PMID)
	fmt.Printf("Authors: %s\n", articleByDOI.AuthorString)
}
