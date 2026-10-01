package main

import (
	"fmt"
	"log"

	"github.com/dictybase/literature"
)

func main() {
	// Create a new OpenAlex client in the "Polite Pool" by providing an
	// email. For OpenAlex Premium use WithOpenAlexAPIKey instead.
	client, err := literature.NewOpenAlexClient(
		literature.WithOpenAlexEmail("you@example.org"),
	)
	if err != nil {
		log.Fatalf("Failed to create OpenAlex client: %v", err)
	}

	const examplePMID = "23842501"
	const exampleOpenAlexID = "W2100837269"

	// Fetch a work by PMID
	fmt.Println("Fetching work by PMID:", examplePMID)
	work, err := client.GetWorkByPMID(examplePMID)
	if err != nil {
		log.Fatalf("Failed to fetch work by PMID: %v", err)
	}
	fmt.Printf("Title: %s\n", work.Title)
	fmt.Printf("OpenAlex ID: %s\n", work.OpenAlexID)
	fmt.Printf("DOI: %s\n", work.DOI)
	fmt.Printf("Cited by: %d\n", work.CitedByCount)
	if len(work.Authors) > 0 {
		fmt.Printf("First author: %s\n", work.Authors[0].Author.DisplayName)
	}
	if len(work.Topics) > 0 {
		fmt.Printf("Topic: %s (Domain: %s)\n",
			work.Topics[0].DisplayName,
			work.Topics[0].Domain.DisplayName,
		)
	}
	if work.OpenAccess.IsOA {
		fmt.Printf("Open access: %s (%s)\n", work.OpenAccess.OAStatus, work.OpenAccess.OAURL)
	}
	fmt.Println()

	// Citation impact metrics
	fmt.Println("Fetching citation metrics for:", exampleOpenAlexID)
	metrics, err := client.GetCitationMetrics(exampleOpenAlexID)
	if err != nil {
		log.Fatalf("Failed to fetch citation metrics: %v", err)
	}
	fmt.Printf("Cited by count: %d\n", metrics.CitedByCount)
	fmt.Printf("Field-Weighted Citation Impact: %.2f\n", metrics.FWCI)
	fmt.Printf("Percentiles (min/max): %.2f / %.2f\n", metrics.PercentileMin, metrics.PercentileMax)
	fmt.Println()

	// Works cited by the article
	fmt.Println("Fetching works cited by:", exampleOpenAlexID)
	referenced, err := client.GetReferencedWorks(exampleOpenAlexID)
	if err != nil {
		log.Fatalf("Failed to fetch referenced works: %v", err)
	}
	fmt.Printf("Found %d referenced works\n", len(referenced))
	for i, ref := range referenced {
		if i >= 5 {
			fmt.Printf("... and %d more\n", len(referenced)-i)
			break
		}
		fmt.Printf("  - %s (%d)\n", ref.Title, ref.PublicationYear)
	}
	fmt.Println()

	// Live stream of works citing the article
	fmt.Println("Fetching citing works for:", exampleOpenAlexID)
	citing, err := client.GetCitingWorks(
		exampleOpenAlexID,
		literature.WithOpenAlexPerPage(5),
		literature.WithOpenAlexSort("cited_by_count:desc"),
	)
	if err != nil {
		log.Fatalf("Failed to fetch citing works: %v", err)
	}
	fmt.Printf("Total citing works: %d\n", citing.TotalCount)
	for _, citingWork := range citing.Results {
		fmt.Printf("  - %s (%d)\n", citingWork.Title, citingWork.PublicationYear)
	}
	fmt.Printf("Next page cursor: %s\n", citing.Cursor)
}
