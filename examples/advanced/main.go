package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cybersiddhu/literature"
)

func main() {
	// Example 1: Create client with custom configuration
	fmt.Println("=== Advanced Client Configuration ===")
	
	customHTTPClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
	}

	client, err := literature.New(
		literature.WithHTTPClient(customHTTPClient),
		literature.WithTimeout(60*time.Second),
		literature.WithUserAgent("MyApp/1.0 (research@example.com)"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 2: Batch article fetching
	fmt.Println("\n=== Batch Article Fetching ===")
	pmids := []string{"33515252", "33515253", "33515254"} // Replace with real PMIDs
	
	articles, err := client.GetArticles(pmids)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
	} else {
		fmt.Printf("Successfully fetched %d articles\n", len(articles))
		for i, article := range articles {
			fmt.Printf("%d. [%s] %s\n", i+1, article.PMID, article.Title)
		}
	}

	// Example 3: Advanced search with pagination
	fmt.Println("\n=== Advanced Search with Pagination ===")
	
	// First page
	firstPage, err := client.Search("machine learning genomics",
		literature.WithLimit(5),
		literature.WithOffset(0),
	)
	if err != nil {
		log.Printf("Error in first search: %v", err)
	} else {
		fmt.Printf("First page: %d results (total: %d)\n", len(firstPage.Articles), firstPage.Total)
	}

	// Second page
	secondPage, err := client.Search("machine learning genomics",
		literature.WithLimit(5),
		literature.WithOffset(5),
	)
	if err != nil {
		log.Printf("Error in second search: %v", err)
	} else {
		fmt.Printf("Second page: %d results\n", len(secondPage.Articles))
	}

	// Example 4: Finding similar articles
	fmt.Println("\n=== Finding Similar Articles ===")
	if len(pmids) > 0 {
		similarArticles, err := client.FindSimilar(pmids[0],
			literature.WithLimit(3),
		)
		if err != nil {
			log.Printf("Error finding similar articles: %v", err)
		} else {
			fmt.Printf("Found %d similar articles to PMID %s:\n", len(similarArticles.Articles), pmids[0])
			for i, article := range similarArticles.Articles {
				fmt.Printf("%d. %s\n", i+1, article.Title)
			}
		}
	}

	// Example 5: Error handling with type checking
	fmt.Println("\n=== Advanced Error Handling ===")
	_, err = client.GetArticle("invalid-pmid")
	if err != nil {
		if litErr, ok := err.(*literature.Error); ok {
			fmt.Printf("Error type: %s\n", litErr.Type)
			fmt.Printf("Error message: %s\n", litErr.Message)
			if litErr.PMID != "" {
				fmt.Printf("Related PMID: %s\n", litErr.PMID)
			}
			
			// Check specific error types
			if litErr.IsType(literature.ErrorTypeInvalidInput) {
				fmt.Println("This was an invalid input error")
			}
		} else {
			fmt.Printf("Unknown error: %v\n", err)
		}
	}

	// Example 6: PDF operations
	fmt.Println("\n=== PDF Operations ===")
	if len(pmids) > 0 {
		pdfInfo, err := client.GetPDF(pmids[0])
		if err != nil {
			log.Printf("Error getting PDF info: %v", err)
		} else {
			fmt.Printf("PDF URL: %s\n", pdfInfo.URL)
			fmt.Printf("Suggested filename: %s\n", pdfInfo.Filename)
		}
	}

	// Example 7: Concurrent operations
	fmt.Println("\n=== Concurrent Operations ===")
	type result struct {
		pmid    string
		article *literature.Article
		err     error
	}

	results := make(chan result, len(pmids))
	
	// Start concurrent fetches
	for _, pmid := range pmids {
		go func(id string) {
			article, err := client.GetArticle(id)
			results <- result{pmid: id, article: article, err: err}
		}(pmid)
	}

	// Collect results
	successCount := 0
	for i := 0; i < len(pmids); i++ {
		res := <-results
		if res.err != nil {
			fmt.Printf("Error fetching PMID %s: %v\n", res.pmid, res.err)
		} else {
			fmt.Printf("Successfully fetched: [%s] %s\n", res.pmid, res.article.Title)
			successCount++
		}
	}
	
	fmt.Printf("Concurrent fetch completed: %d/%d successful\n", successCount, len(pmids))
}