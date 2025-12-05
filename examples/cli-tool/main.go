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
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictybase/literature"
	"github.com/urfave/cli/v2"
)

// Context is the base context for the CLI tool.
type Context struct {
	Identifier string
	OutputFile string
	Logger     *log.Logger
}

// WithEuropeClient adds the EuropePMC client to the context.
type WithEuropeClient struct {
	Context
	Europe *literature.EuropePMCClient
}

// WithPubMedClient adds the PubMed client to the context.
type WithPubMedClient struct {
	WithEuropeClient
	PubMed *literature.Client
}

var (
	// SetEuropeClient sets the EuropePMC client in the context.
	SetEuropeClient = F.Curry2(
		func(epc *literature.EuropePMCClient, ctx Context) WithEuropeClient {
			return WithEuropeClient{Context: ctx, Europe: epc}
		},
	)

	// SetPubMedClient sets the PubMed client in the context.
	SetPubMedClient = F.Curry2(
		func(epc *literature.Client, ctx WithEuropeClient) WithPubMedClient {
			return WithPubMedClient{WithEuropeClient: ctx, PubMed: epc}
		},
	)
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

func run(ctx *cli.Context) error {
	identifier := ctx.Args().First()
	if identifier == "" {
		return cli.Exit("Please provide a PMID or DOI", 1)
	}

	outputFile := ctx.String("output")
	logger := log.Default()
	logger.SetOutput(os.Stderr)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Construct the program
	return F.Pipe5(
		IOE.Do[error](Context{
			Identifier: identifier,
			OutputFile: outputFile,
			Logger:     logger,
		}),
		IOE.Bind(SetEuropeClient, createEuropeClient),
		IOE.Bind(SetPubMedClient, createPubMedClient),
		IOE.Chain(fetchAndProcess),
		ToEither,
		E.Fold(
			F.Identity[error],
			F.Constant1[any, error](nil),
		),
	)
}

func fetchAndProcess(
	ctx WithPubMedClient,
) IOE.IOEither[error, any] {
	// EuropePMC Flow
	europeFlow := F.Pipe2(
		IOE.Of[error](ctx),
		IOE.Chain(F.Ternary(isDOI, europeByDOI, europeByPMID)),
		IOE.Chain(processEuropeArticle(ctx)),
	)

	// PubMed Flow (Fallback)
	pubMedFlow := func() IOE.IOEither[error, any] {
		return F.Pipe3(
			IOE.Of[error](ctx),
			IOE.ChainFirstIOK[error](
				IO.Logger[WithPubMedClient](
					ctx.Logger,
				)(
					"Not found in EuropePMC. Trying PubMed...",
				),
			),
			IOE.Chain(resolvePMID),
			IOE.Chain(processPubMedFlow(ctx)),
		)
	}

	// Combine with Alt
	return IOE.MonadAlt(europeFlow, pubMedFlow)
}

func processEuropeArticle(
	ctx WithPubMedClient,
) func(*literature.EuropePMCArticle) IOE.IOEither[error, any] {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
		return F.Pipe1(
			logEuropeArticle(ctx)(article),
			IOE.Chain(
				func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
					if article.HasPDF {
						return F.Pipe1(
							processEuropePDF(ctx)(article),
							IOE.ChainFirst(
								func(_ any) IOE.IOEither[error, any] {
									return logInfo(
										ctx,
										"PDF available via EuropePMC.",
									)
								},
							),
						)
					}
					return processEuropeFallback(ctx)(article)
				},
			),
		)
	}
}

func processEuropePDF(
	ctx WithPubMedClient,
) func(*literature.EuropePMCArticle) IOE.IOEither[error, any] {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
		if article.PMID == "" {
			return logInfo(
				ctx,
				"Warning: No PMID, cannot reliably fetch PDF.",
			)
		}
		downloadOp := downloadEuropePDF(
			ctx.Europe,
			article.PMID,
			ctx.OutputFile,
		)
		return IOE.MonadMap(
			downloadOp,
			func(string) any { return nil },
		)
	}
}

func processEuropeFallback(
	ctx WithPubMedClient,
) func(*literature.EuropePMCArticle) IOE.IOEither[error, any] {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
		return F.Pipe1(
			logInfo(
				ctx,
				"PDF NOT available via EuropePMC. Trying PubMed for PDF...",
			),
			IOE.Chain(func(_ any) IOE.IOEither[error, any] {
				if article.PMID == "" {
					return logInfo(
						ctx,
						"No PMID available to query PubMed for PDF.",
					)
				}
				return downloadPubMedPDF(
					ctx.PubMed,
					article.PMID,
					ctx.OutputFile,
				)
			}),
		)
	}
}

func processPubMedFlow(
	ctx WithPubMedClient,
) func(string) IOE.IOEither[error, any] {
	return func(pmid string) IOE.IOEither[error, any] {
		return F.Pipe1(
			fetchPubMedArticle(pmid)(ctx),
			IOE.Chain(
				func(article *literature.Article) IOE.IOEither[error, any] {
					return F.Pipe1(
						logPubMedArticle(ctx)(article),
						IOE.Chain(
							func(art *literature.Article) IOE.IOEither[error, any] {
								return downloadPubMedPDF(
									ctx.PubMed,
									art.PMID,
									ctx.OutputFile,
								)
							},
						),
					)
				},
			),
		)
	}
}

// --- Wrappers ---

func createEuropeClient(
	_ Context,
) IOE.IOEither[error, *literature.EuropePMCClient] {
	return IOE.TryCatchError(func() (*literature.EuropePMCClient, error) {
		return literature.NewEuropePMCClient()
	})
}

func createPubMedClient(
	_ WithEuropeClient,
) IOE.IOEither[error, *literature.Client] {
	return IOE.TryCatchError(func() (*literature.Client, error) {
		return literature.New()
	})
}

func ToEither[A any](ioe IOE.IOEither[error, A]) E.Either[error, A] {
	return ioe()
}

func logInfo(ctx WithPubMedClient, msg string) IOE.IOEither[error, any] {
	return IOE.Of[error, any](func() any {
		ctx.Logger.Info(msg)
		return nil
	})
}

func logEuropeArticle(
	ctx WithPubMedClient,
) func(*literature.EuropePMCArticle) IOE.IOEither[error, *literature.EuropePMCArticle] {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, *literature.EuropePMCArticle] {
		return IOE.ChainFirst(
			func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
				return IOE.Of[error, any](func() any {
					ctx.Logger.Info(
						"Article Details (EuropePMC)",
						"title", article.Title,
						"authors", article.AuthorString,
						"pmid", article.PMID,
						"doi", article.DOI,
					)
					return nil
				})
			},
		)(
			IOE.Of[error](article),
		)
	}
}

func logPubMedArticle(
	ctx WithPubMedClient,
) func(*literature.Article) IOE.IOEither[error, *literature.Article] {
	return func(article *literature.Article) IOE.IOEither[error, *literature.Article] {
		return IOE.ChainFirst(
			func(article *literature.Article) IOE.IOEither[error, any] {
				return IOE.Of[error, any](func() any {
					ctx.Logger.Info(
						"Article Details (PubMed)",
						"title", article.Title,
						"pmid", article.PMID,
						"doi", article.DOI,
					)
					return nil
				})
			},
		)(IOE.Of[error](article))
	}
}

func isDOI(ctx WithPubMedClient) bool {
	return F.Pipe1(
		ctx.Identifier,
		P.Or(S.Includes("/"))(S.HasPrefix("10.")),
	)
}

func europeByDOI(
	ctx WithPubMedClient,
) IOE.IOEither[error, *literature.EuropePMCArticle] {
	return IOE.TryCatchError(
		func() (*literature.EuropePMCArticle, error) {
			return ctx.Europe.GetArticleByDOI(ctx.Identifier)
		},
	)
}

func europeByPMID(
	ctx WithPubMedClient,
) IOE.IOEither[error, *literature.EuropePMCArticle] {
	return IOE.TryCatchError(
		func() (*literature.EuropePMCArticle, error) {
			return ctx.Europe.GetArticle(ctx.Identifier)
		},
	)
}

func resolvePMID(ctx WithPubMedClient) IOE.IOEither[error, string] {
	identifier := ctx.Identifier
	return IOE.TryCatchError(func() (string, error) {
		isDOI := strings.Contains(identifier, "/") ||
			strings.HasPrefix(identifier, "10.")
		if !isDOI {
			return identifier, nil
		}
		searchResults, err := ctx.PubMed.Search(identifier)
		if err != nil || searchResults.Total == 0 {
			return "", fmt.Errorf(
				"article not found in PubMed via DOI: %s",
				identifier,
			)
		}
		if len(searchResults.Articles) > 0 {
			return searchResults.Articles[0].PMID, nil
		}
		return "", fmt.Errorf(
			"DOI resolved but no article details returned",
		)
	})
}

var fetchPubMedArticle = F.Curry2(
	func(pmid string, ctx WithPubMedClient) IOE.IOEither[error, *literature.Article] {
		return IOE.TryCatchError(func() (*literature.Article, error) {
			return ctx.PubMed.GetArticle(pmid)
		})
	},
)

func downloadEuropePDF(
	client *literature.EuropePMCClient,
	pmid, customName string,
) IOE.IOEither[error, string] {
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

func downloadPubMedPDF(
	client *literature.Client,
	pmid, customName string,
) IOE.IOEither[error, any] {
	return IOE.TryCatchError(func() (any, error) {
		hasPDF, err := client.HasPDF(pmid)
		if err != nil {
			//nolint:nilerr // Not a fatal error for the CLI flow, just no PDF
			return nil, nil
		}
		if !hasPDF {
			return nil, nil
		}
		filename := getFilename(pmid, customName)
		fmt.Printf("Downloading PDF from PubMed to %s...\n", filename)
		err = client.DownloadPDF(pmid, filename)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to download PDF from PubMed: %w",
				err,
			)
		}
		fmt.Println("PDF downloaded successfully.")
		return nil, nil
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
