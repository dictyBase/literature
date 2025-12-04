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
		Usage: "Fetch article metadata and download PDF",
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

	// Create a new EuropePMC client
	client, err := literature.NewEuropePMCClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	var article *literature.EuropePMCArticle
	isDOI := strings.Contains(id, "/") || strings.HasPrefix(id, "10.")

	if isDOI {
		fmt.Printf("Fetching by DOI: %s...\n", id)
		article, err = client.GetArticleByDOI(id)
	} else {
		fmt.Printf("Fetching by PMID: %s...\n", id)
		article, err = client.GetArticle(id)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch article: %w", err)
	}

	printArticle(article)

	// Check for PDF availability
	if article.HasPDF {
		fmt.Println("\nPDF is available.")
		if article.PMID == "" {
			fmt.Println("Warning: Article has PDF flag but no PMID, cannot fetch PDF URLs via this client.")
			return nil
		}
		return downloadPDF(client, article.PMID, c.String("output"))
	} else {
		fmt.Println("\nNo full text PDF available via EuropePMC.")
	}

	return nil
}

func printArticle(a *literature.EuropePMCArticle) {
	fmt.Println("\n=== Article Details ===")
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

func downloadPDF(client *literature.EuropePMCClient, pmid, customName string) error {
	urls, err := client.GetPDFURLs(pmid)
	if err != nil {
		return fmt.Errorf("failed to get PDF URLs: %w", err)
	}

	if len(urls) == 0 {
		return fmt.Errorf("no PDF URLs returned")
	}

	// Use the first available URL
	pdfURL := urls[0].URL
	fmt.Printf("Downloading PDF from: %s\n", pdfURL)

	// Determine filename
	filename := customName
	if filename == "" {
		filename = fmt.Sprintf("%s.pdf", pmid)
	}

	// Download the file
	resp, err := http.Get(pdfURL)
	if err != nil {
		return fmt.Errorf("failed to initiate download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	fmt.Printf("PDF saved to: %s\n", filename)
	return nil
}
