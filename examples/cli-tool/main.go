package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	IOEF "github.com/IBM/fp-go/v2/ioeither/file"
	IOEH "github.com/IBM/fp-go/v2/ioeither/http"
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

type DownloadContext struct {
	WithPubMedClient
	PMID       string
	PDFURL     string
	TargetFile string
}

type DownloadFlow = func(*literature.EuropePMCArticle) IOE.IOEither[error, any]

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

	pubClientLogger = F.Curry2(
		func(msg string, ctx WithPubMedClient) IO.IO[WithPubMedClient] {
			return func() WithPubMedClient {
				ctx.Logger.Print(msg)
				return ctx
			}
		},
	)

	logEuropeArticle = F.Curry2(
		func(ctx WithPubMedClient, article *literature.EuropePMCArticle) IO.IO[*literature.EuropePMCArticle] {
			return func() *literature.EuropePMCArticle {
				ctx.Logger.Println(
					"Article Details (EuropePMC)",
					"title", article.Title,
					"authors", article.AuthorString,
					"pmid", article.PMID,
					"doi", article.DOI,
				)
				return nil
			}
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
	europeFlow := F.Pipe3(
		IOE.Of[error](ctx),
		IOE.Chain(F.Ternary(isDOI, europeByDOI, europeByPMID)),
		IOE.ChainFirstIOK[error](logEuropeArticle(ctx)),
		IOE.Chain(
			F.Ternary(
				hasEuropePDF,
				downloadEuropePDF(ctx),
				fallbackEurope(ctx),
			),
		),
	)

	// PubMed Flow (Fallback)
	pubMedFlow := func() IOE.IOEither[error, any] {
		return F.Pipe3(
			IOE.Of[error](ctx),
			IOE.ChainFirstIOK[error](
				pubClientLogger("Not found in EuropePMC. Trying PubMed..."),
			),
			IOE.Chain(resolvePMID),
			IOE.Chain(processPubMedFlow(ctx)),
		)
	}

	// Combine with Alt
	return IOE.MonadAlt(europeFlow, pubMedFlow)
}

func hasEuropePDF(a *literature.EuropePMCArticle) bool {
	return a.HasPDF
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
								return F.Pipe4(
									IOE.Of[error](
										DownloadContext{
											WithPubMedClient: ctx,
											PMID:             art.PMID,
										},
									),
									IOE.Chain(checkPubMedAvailability),
									IOE.Map[error](setTargetFilename),
									IOE.Chain(downloadFromPubMed),
									IOE.Map[error](
										F.Constant1[DownloadContext, any](nil),
									),
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

func logPubMedArticle(
	ctx WithPubMedClient,
) func(*literature.Article) IOE.IOEither[error, *literature.Article] {
	return func(article *literature.Article) IOE.IOEither[error, *literature.Article] {
		return IOE.ChainFirst(
			func(article *literature.Article) IOE.IOEither[error, any] {
				return IOE.Of[error, any](func() any {
					ctx.Logger.Println(
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

func searchPubMed(
	ctx WithPubMedClient,
) IOE.IOEither[error, *literature.SearchResult] {
	return IOE.TryCatchError(
		func() (*literature.SearchResult, error) {
			return ctx.PubMed.Search(ctx.Identifier)
		},
	)
}

func extractFirstPMID(
	result *literature.SearchResult,
) IOE.IOEither[error, string] {
	return IOE.TryCatchError(func() (string, error) {
		if result.Total == 0 {
			return "", fmt.Errorf("article not found in PubMed")
		}
		if len(result.Articles) > 0 {
			return result.Articles[0].PMID, nil
		}
		return "", fmt.Errorf(
			"DOI resolved but no article details returned",
		)
	})
}

func resolveDirectly(ctx WithPubMedClient) IOE.IOEither[error, string] {
	return IOE.Of[error](ctx.Identifier)
}

var resolveFromPubMedSearch = F.Flow2(
	searchPubMed,
	IOE.Chain(extractFirstPMID),
)

func resolvePMID(ctx WithPubMedClient) IOE.IOEither[error, string] {
	return F.Pipe1(
		IOE.Of[error](ctx),
		IOE.Chain(
			F.Ternary(
				isDOI,
				resolveFromPubMedSearch,
				resolveDirectly,
			),
		),
	)
}

var fetchPubMedArticle = F.Curry2(
	func(pmid string, ctx WithPubMedClient) IOE.IOEither[error, *literature.Article] {
		return IOE.TryCatchError(func() (*literature.Article, error) {
			return ctx.PubMed.GetArticle(pmid)
		})
	},
)

func downloadEuropePDF(
	ctx WithPubMedClient,
) DownloadFlow {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
		return F.Pipe4(
			IOE.Of[error](
				DownloadContext{WithPubMedClient: ctx, PMID: article.PMID},
			),
			IOE.Chain(getPDFURLs),
			IOE.Map[error](setTargetFilename),
			IOE.Chain(downloadPDF),
			IOE.Map[error](F.Constant1[DownloadContext, any](nil)),
		)
	}
}

func getFilename(pmid, customName string) string {
	if customName != "" {
		return customName
	}
	return fmt.Sprintf("%s.pdf", pmid)
}

func getPDFURLs(ctx DownloadContext) IOE.IOEither[error, DownloadContext] {
	return IOE.TryCatchError(func() (DownloadContext, error) {
		urls, err := ctx.Europe.GetPDFURLs(ctx.PMID)
		if err != nil {
			return ctx, err
		}
		ctx.PDFURL = urls[0].URL
		return ctx, nil
	})
}

func setTargetFilename(ctx DownloadContext) DownloadContext {
	ctx.TargetFile = getFilename(ctx.PMID, ctx.OutputFile)
	return ctx
}

func downloadPDF(ctx DownloadContext) IOE.IOEither[error, DownloadContext] {
	return F.Pipe3(
		ctx.PDFURL,
		IOEH.MakeGetRequest,
		IOEH.ReadAll(IOEH.MakeClient(http.DefaultClient)),
		IOE.Chain(func(data []byte) IOE.IOEither[error, DownloadContext] {
			return F.Pipe1(
				IOEF.WriteFile(ctx.TargetFile, 0644)(data),
				IOE.Map[error](F.Constant1[[]byte](ctx)),
			)
		}),
	)
}

func checkPubMedAvailability(
	ctx DownloadContext,
) IOE.IOEither[error, DownloadContext] {
	return IOE.TryCatchError(func() (DownloadContext, error) {
		hasPDF, err := ctx.PubMed.HasPDF(ctx.PMID)
		if err != nil {
			return ctx, err
		}
		if !hasPDF {
			return ctx, fmt.Errorf("PDF not available in PubMed")
		}
		return ctx, nil
	})
}

func downloadFromPubMed(
	ctx DownloadContext,
) IOE.IOEither[error, DownloadContext] {
	return IOE.TryCatchError(func() (DownloadContext, error) {
		err := ctx.PubMed.DownloadPDF(ctx.PMID, ctx.TargetFile)
		if err != nil {
			return ctx, fmt.Errorf(
				"failed to download PDF from PubMed: %w",
				err,
			)
		}
		return ctx, nil
	})
}

func fallbackEurope(
	ctx WithPubMedClient,
) DownloadFlow {
	return func(article *literature.EuropePMCArticle) IOE.IOEither[error, any] {
		return F.Pipe4(
			IOE.Of[error](DownloadContext{
				WithPubMedClient: ctx,
				PMID:             article.PMID,
			}),
			IOE.Chain(checkPubMedAvailability),
			IOE.Map[error](setTargetFilename),
			IOE.Chain(downloadFromPubMed),
			IOE.Map[error](F.Constant1[DownloadContext, any](nil)),
		)
	}
}
