package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dictybase/literature"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "lit-cli",
		Usage: "Fetch article metadata and download PDF (EuropePMC with PubMed fallback)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output filename for PDF",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	id := c.Args().First()
	if id == "" {
		return cli.Exit("Please provide a PMID or DOI", 1)
	}

	// Initialize Clients
	europeClient, err := literature.NewEuropePMCClient()
	if err != nil {
		return fmt.Errorf("failed to create EuropePMC client: %w", err)
	}

	pubmedClient, err := literature.New()
	if err != nil {
		return fmt.Errorf("failed to create PubMed client: %w", err)
	}

	isDOI := strings.Contains(id, "/") || strings.HasPrefix(id, "10.")
	outputFile := c.String("output")

	// --- Attempt 1: EuropePMC ---
	fmt.Println("Checking EuropePMC...")
	var article *literature.EuropePMCArticle
	var europeErr error

	if isDOI {
		article, europeErr = europeClient.GetArticleByDOI(id)
	} else {
		article, europeErr = europeClient.GetArticle(id)
	}

	if europeErr == nil && article != nil {
		// Found in EuropePMC
		printEuropePMCArticle(article)

		if article.HasPDF {
			fmt.Println("\nPDF available via EuropePMC.")
			if article.PMID == "" {
				fmt.Println("Warning: No PMID, cannot reliably fetch PDF.")
				return nil
			}
			return downloadEuropePMCPDF(europeClient, article.PMID, outputFile)
		} else {
			fmt.Println("\nPDF NOT available via EuropePMC. Trying PubMed for PDF...")
			// Fallback for PDF only
			pmid := article.PMID
			if pmid == "" {
				fmt.Println("No PMID available to query PubMed for PDF.")
				return nil
			}
			return tryPubMedPDF(pubmedClient, pmid, outputFile)
		}
	}

	// --- Attempt 2: PubMed (Fallback) ---
	fmt.Printf("\nNot found in EuropePMC (%v). Trying PubMed...\n", europeErr)

	var pubmedArticle *literature.Article
	var pmid string

	if isDOI {
		// Search for PMID using DOI
		fmt.Printf("Resolving DOI %s in PubMed...\n", id)
		searchResults, err := pubmedClient.Search(id)
		if err != nil || searchResults.Total == 0 {
			return fmt.Errorf("article not found in PubMed via DOI: %s", id)
		}
		// Assuming the first result is the correct one
		if len(searchResults.Articles) > 0 {
			pmid = searchResults.Articles[0].PMID
			pubmedArticle = searchResults.Articles[0]
		} else {
			// Fallback if search returns count but no detailed articles in list (unlikely with current logic)
			return fmt.Errorf("DOI resolved but no article details returned")
		}
	} else {
		pmid = id
		pubmedArticle, err = pubmedClient.GetArticle(pmid)
		if err != nil {
			return fmt.Errorf("article not found in PubMed: %w", err)
		}
	}

	if pubmedArticle != nil {
		printPubMedArticle(pubmedArticle)
		return tryPubMedPDF(pubmedClient, pubmedArticle.PMID, outputFile)
	}

	return fmt.Errorf("article not found in either service")
}

func printEuropePMCArticle(a *literature.EuropePMCArticle) {
	fmt.Println("\n=== Article Details (EuropePMC) ===")
	fmt.Printf("Title:    %s\n", a.Title)
	fmt.Printf("Authors:  %s\n", a.AuthorString)
	if a.PMID != "" {
		fmt.Printf("PMID:     %s\n", a.PMID)
	}
	if a.DOI != "" {
		fmt.Printf("DOI:      %s\n", a.DOI)
	}
	fmt.Println("\n--- Abstract ---")
	if a.Abstract != "" {
		fmt.Println(a.Abstract)
	} else {
		fmt.Println("(No abstract available)")
	}
	fmt.Println("----------------")
}

func printPubMedArticle(a *literature.Article) {
	fmt.Println("\n=== Article Details (PubMed) ===")
	fmt.Printf("Title:    %s\n", a.Title)
	fmt.Printf("Authors:  ")
	for i, author := range a.Authors {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(author.FullName)
	}
	fmt.Println()
	fmt.Printf("PMID:     %s\n", a.PMID)
	if a.DOI != "" {
		fmt.Printf("DOI:      %s\n", a.DOI)
	}
	fmt.Println("\n--- Abstract ---")
	if a.Abstract != "" {
		fmt.Println(a.Abstract)
	} else {
		fmt.Println("(No abstract available)")
	}
	fmt.Println("----------------")
}

func downloadEuropePMCPDF(client *literature.EuropePMCClient, pmid, customName string) error {
	urls, err := client.GetPDFURLs(pmid)
	if err != nil {
		return fmt.Errorf("failed to get PDF URLs: %w", err)
	}

	if len(urls) == 0 {
		return fmt.Errorf("no PDF URLs returned")
	}

	// Use the first available URL
	pdfURL := urls[0].URL
	fmt.Printf("Downloading PDF from EuropePMC: %s\n", pdfURL)

	filename := getFilename(pmid, customName)
	return downloadFile(pdfURL, filename)
}

func tryPubMedPDF(client *literature.Client, pmid, customName string) error {
	fmt.Println("Checking PDF availability in PubMed...")
	hasPDF, err := client.HasPDF(pmid)
	if err != nil {
		fmt.Printf("Error checking PubMed PDF: %v\n", err)
		return nil
	}

	if !hasPDF {
		fmt.Println("PDF not available in PubMed.")
		return nil
	}

	filename := getFilename(pmid, customName)
	fmt.Printf("Downloading PDF from PubMed to %s...\n", filename)
	err = client.DownloadPDF(pmid, filename)
	if err != nil {
		return fmt.Errorf("failed to download PDF from PubMed: %w", err)
	}
	fmt.Println("PDF downloaded successfully.")
	return nil
}

func getFilename(pmid, customName string) string {
	if customName != "" {
		return customName
	}
	return fmt.Sprintf("%s.pdf", pmid)
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to initiate download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	
	fmt.Printf("PDF saved to: %s\n", filepath)
	return nil
}
