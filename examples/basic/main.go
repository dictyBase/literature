package main

import (
	"fmt"
	"log"

	"github.com/dictybase/literature"
)

func main() {
	// Create a new literature client with default settings
	client, err := literature.New()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: Fetch a single article by PMID
	fmt.Println("=== Fetching Article ===")
	pmid := "12345678" // Replace with a real PMID
	article, err := client.GetArticle(pmid)
	if err != nil {
		log.Printf("Error fetching article: %v", err)
	} else {
		fmt.Printf("Title: %s\n", article.Title)
		fmt.Printf("Journal: %s\n", article.Journal)
		fmt.Printf("Authors: ")
		for i, author := range article.Authors {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(author.FullName)
		}
		fmt.Println()
		if article.Abstract != "" {
			fmt.Printf("Abstract: %s...\n", truncateString(article.Abstract, 100))
		}
	}

	// Example 2: Search for articles
	fmt.Println("\n=== Searching Articles ===")
	searchResults, err := client.Search("cancer treatment")
	if err != nil {
		log.Printf("Error searching: %v", err)
	} else {
		fmt.Printf("Found %d articles for query 'cancer treatment'\n", searchResults.Total)
		fmt.Printf("Showing %d results:\n", len(searchResults.Articles))
		for i, article := range searchResults.Articles {
			fmt.Printf("%d. %s\n", i+1, article.Title)
		}
	}

	// Example 3: Check for PDF availability
	fmt.Println("\n=== Checking PDF Availability ===")
	hasPDF, err := client.HasPDF(pmid)
	if err != nil {
		log.Printf("Error checking PDF: %v", err)
	} else {
		fmt.Printf("PDF available for PMID %s: %v\n", pmid, hasPDF)
	}
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
