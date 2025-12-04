package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
	"github.com/urfave/cli/v2"
)

// Clients holds the API clients.
type Clients struct {
	Europe *literature.EuropePMCClient
	PubMed *literature.Client
}

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

func run(ctx *cli.Context) error {
	identifier := ctx.Args().First()
	if identifier == "" {
		return cli.Exit("Please provide a PMID or DOI", 1)
	}

	outputFile := ctx.String("output")

	// Construct the program
	program := F.Pipe1(
		createClients(),
		IOE.Chain(func(clients Clients) IOE.IOEither[error, any] {
			return fetchAndProcess(clients, identifier, outputFile)
		}),
	)

	// Execute the program
	return E.Fold(
		F.Identity[error],
		F.Constant1[any, error](nil),
	)(program())
}

// --- Logic Flow ---

func fetchAndProcess(clients Clients, identifier, outputFile string) IOE.IOEither[error, any] {
	// EuropePMC Flow
	europeFlow := F.Pipe1(
		fetchEuropeArticle(clients.Europe, identifier),
		IOE.Chain(processEuropeArticle(clients, outputFile)),
	)

	// PubMed Flow (Fallback)
	pubMedFlow := func() IOE.IOEither[error, any] {
		fmt.Printf("\nNot found in EuropePMC. Trying PubMed...\n")
		return F.Pipe1(
			resolvePMID(clients.PubMed, identifier),
			IOE.Chain(processPubMedFlow(clients, outputFile)),
		)
	}

	// Combine with Alt
	return IOE.MonadAlt(europeFlow, pubMedFlow)
}

func processEuropeArticle(clients Clients, outputFile string) func(*literature.EuropePMCArticle) IOE.IOEither[error, any] {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
		return F.Pipe1(
			printEuropeDetails(article),
			IOE.Chain(func(a *literature.EuropePMCArticle) IOE.IOEither[error, any] {
				if a.HasPDF {
					fmt.Println("\nPDF available via EuropePMC.")
					if a.PMID == "" {
						fmt.Println("Warning: No PMID, cannot reliably fetch PDF.")
						return IOE.Of[error, any](nil)
					}
					// Transform IOEither[error, string] to IOEither[error, any]
					downloadOp := downloadEuropePDF(clients.Europe, a.PMID, outputFile)
					return IOE.MonadMap(downloadOp, func(string) any { return nil })
				}
				// Partial Fallback
				fmt.Println("\nPDF NOT available via EuropePMC. Trying PubMed for PDF...")
				if a.PMID == "" {
					fmt.Println("No PMID available to query PubMed for PDF.")
					return IOE.Of[error, any](nil)
				}
				return downloadPubMedPDF(clients.PubMed, a.PMID, outputFile)
			}),
		)
	}
}

func processPubMedFlow(clients Clients, outputFile string) func(string) IOE.IOEither[error, any] {
	return func(pmid string) IOE.IOEither[error, any] {
		return F.Pipe1(
			fetchPubMedArticle(clients.PubMed, pmid),
			IOE.Chain(func(article *literature.Article) IOE.IOEither[error, any] {
				return F.Pipe1(
					printPubMedDetails(article),
					IOE.Chain(func(a *literature.Article) IOE.IOEither[error, any] {
						return downloadPubMedPDF(clients.PubMed, a.PMID, outputFile)
					}),
				)
			}),
		)
	}
}

// --- Wrappers ---

func createClients() IOE.IOEither[error, Clients] {
	return IOE.TryCatchError(func() (Clients, error) {
		europeClient, err := literature.NewEuropePMCClient()
		if err != nil {
			return Clients{}, fmt.Errorf("failed to create EuropePMC client: %w", err)
		}
		pubMedClient, err := literature.New()
		if err != nil {
			return Clients{}, fmt.Errorf("failed to create PubMed client: %w", err)
		}
		return Clients{Europe: europeClient, PubMed: pubMedClient}, nil
	})
}

func fetchEuropeArticle(client *literature.EuropePMCClient, identifier string) IOE.IOEither[error, *literature.EuropePMCArticle] {
	return IOE.TryCatchError(func() (*literature.EuropePMCArticle, error) {
		fmt.Println("Checking EuropePMC...")
		isDOI := strings.Contains(identifier, "/") || strings.HasPrefix(identifier, "10.")
		if isDOI {
			return client.GetArticleByDOI(identifier)
		}
		return client.GetArticle(identifier)
	})
}

func resolvePMID(client *literature.Client, identifier string) IOE.IOEither[error, string] {
	return IOE.TryCatchError(func() (string, error) {
		isDOI := strings.Contains(identifier, "/") || strings.HasPrefix(identifier, "10.")
		if !isDOI {
			return identifier, nil
		}
		fmt.Printf("Resolving DOI %s in PubMed...\n", identifier)
		searchResults, err := client.Search(identifier)
		if err != nil || searchResults.Total == 0 {
			return "", fmt.Errorf("article not found in PubMed via DOI: %s", identifier)
		}
		if len(searchResults.Articles) > 0 {
			return searchResults.Articles[0].PMID, nil
		}
		return "", fmt.Errorf("DOI resolved but no article details returned")
	})
}

func fetchPubMedArticle(client *literature.Client, pmid string) IOE.IOEither[error, *literature.Article] {
	return IOE.TryCatchError(func() (*literature.Article, error) {
		return client.GetArticle(pmid)
	})
}

func downloadEuropePDF(client *literature.EuropePMCClient, pmid, customName string) IOE.IOEither[error, string] {
	return IOE.TryCatchError(func() (string, error) {
		urls, err := client.GetPDFURLs(pmid)
		if err != nil {
			return "", err
		}
		if len(urls) == 0 {
			return "", fmt.Errorf("no PDF URLs returned")
		}
		pdfURL := urls[0].URL
		fmt.Printf("Downloading PDF from EuropePMC: %s\n", pdfURL)
		filename := getFilename(pmid, customName)
		return filename, downloadFile(pdfURL, filename)
	})
}

func downloadPubMedPDF(client *literature.Client, pmid, customName string) IOE.IOEither[error, any] {
	return IOE.TryCatchError(func() (any, error) {
		fmt.Println("Checking PDF availability in PubMed...")
		hasPDF, err := client.HasPDF(pmid)
		if err != nil {
			fmt.Printf("Error checking PubMed PDF: %v\n", err)
			return nil, nil // Not a fatal error for the CLI flow, just no PDF
		}
		if !hasPDF {
			fmt.Println("PDF not available in PubMed.")
			return nil, nil
		}
		filename := getFilename(pmid, customName)
		fmt.Printf("Downloading PDF from PubMed to %s...\n", filename)
		err = client.DownloadPDF(pmid, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to download PDF from PubMed: %w", err)
		}
		fmt.Println("PDF downloaded successfully.")
		return nil, nil
	})
}

// --- Utilities ---

func printEuropeDetails(article *literature.EuropePMCArticle) IOE.IOEither[error, *literature.EuropePMCArticle] {
	return IOE.TryCatchError(func() (*literature.EuropePMCArticle, error) {
		fmt.Println("\n=== Article Details (EuropePMC) ===")
		fmt.Printf("Title:    %s\n", article.Title)
		fmt.Printf("Authors:  %s\n", article.AuthorString)
		if article.PMID != "" {
			fmt.Printf("PMID:     %s\n", article.PMID)
		}
		if article.DOI != "" {
			fmt.Printf("DOI:      %s\n", article.DOI)
		}
		fmt.Println("\n--- Abstract ---")
		if article.Abstract != "" {
			fmt.Println(article.Abstract)
		} else {
			fmt.Println("(No abstract available)")
		}
		fmt.Println("----------------")
		return article, nil
	})
}

func printPubMedDetails(article *literature.Article) IOE.IOEither[error, *literature.Article] {
	return IOE.TryCatchError(func() (*literature.Article, error) {
		fmt.Println("\n=== Article Details (PubMed) ===")
		fmt.Printf("Title:    %s\n", article.Title)
		fmt.Printf("Authors:  ")
		for i, author := range article.Authors {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(author.FullName)
		}
		fmt.Println()
		fmt.Printf("PMID:     %s\n", article.PMID)
		if article.DOI != "" {
			fmt.Printf("DOI:      %s\n", article.DOI)
		}
		fmt.Println("\n--- Abstract ---")
		if article.Abstract != "" {
			fmt.Println(article.Abstract)
		} else {
			fmt.Println("(No abstract available)")
		}
		fmt.Println("----------------")
		return article, nil
	})
}

func getFilename(pmid, customName string) string {
	if customName != "" {
		return customName
	}
	return fmt.Sprintf("%s.pdf", pmid)
}

func downloadFile(url, filepath string) error {
	// nosemgrep: go.lang.security.audit.net.http.request.http-request-variable-url
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
