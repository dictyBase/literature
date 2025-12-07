package main

import (
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

func europeByDOI(
	ctx WithPubMedClient,
) IOE.IOEither[error, WithEuropePMCArticle] {
	return F.Pipe1(
		IOE.TryCatchError(
			func() (*literature.EuropePMCArticle, error) {
				return ctx.Europe.GetArticleByDOI(ctx.Identifier)
			},
		),
		IOE.Map[error](
			func(a *literature.EuropePMCArticle) WithEuropePMCArticle {
				return WithEuropePMCArticle{
					WithPubMedClient: ctx,
					Article:          a,
				}
			},
		),
	)
}

func europeByPMID(
	ctx WithPubMedClient,
) IOE.IOEither[error, WithEuropePMCArticle] {
	return F.Pipe1(
		IOE.TryCatchError(
			func() (*literature.EuropePMCArticle, error) {
				return ctx.Europe.GetArticle(ctx.Identifier)
			},
		),
		IOE.Map[error](
			func(a *literature.EuropePMCArticle) WithEuropePMCArticle {
				return WithEuropePMCArticle{
					WithPubMedClient: ctx,
					Article:          a,
				}
			},
		),
	)
}

func hasEuropePDF(ctx WithEuropePMCArticle) bool {
	return ctx.Article.HasPDF
}

func getPDFURLs(ctx DownloadContext) IOE.IOEither[error, DownloadContext] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([]literature.EuropePMCFullTextURL, error) {
			return ctx.Europe.GetPDFURLs(ctx.PMID)
		}),
		IOE.Map[error](
			func(urls []literature.EuropePMCFullTextURL) DownloadContext {
				ctx.PDFURL = urls[0].URL
				return ctx
			},
		),
	)
}

func downloadEuropePDF(
	ctx WithEuropePMCArticle,
) IOE.IOEither[error, any] {
	return F.Pipe4(
		IOE.Of[error](
			DownloadContext{
				WithPubMedClient: ctx.WithPubMedClient,
				PMID:             ctx.Article.PMID,
			},
		),
		IOE.Chain(getPDFURLs),
		IOE.Map[error](setTargetFilename),
		IOE.Chain(downloadPDF),
		IOE.Map[error](F.Constant1[DownloadContext, any](nil)),
	)
}

func fallbackEurope(
	ctx WithEuropePMCArticle,
) IOE.IOEither[error, any] {
	return F.Pipe4(
		IOE.Of[error](DownloadContext{
			WithPubMedClient: ctx.WithPubMedClient,
			PMID:             ctx.Article.PMID,
		}),
		IOE.Chain(checkPubMedAvailability),
		IOE.Map[error](setTargetFilename),
		IOE.Chain(downloadFromPubMed),
		IOE.Map[error](F.Constant1[DownloadContext, any](nil)),
	)
}
