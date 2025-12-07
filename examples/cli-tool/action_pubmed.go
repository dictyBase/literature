package main

import (
	"fmt"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

func fetchPubMedArticle(
	ctx WithPubMedClient,
) IOE.IOEither[error, WithPubMedArticle] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.Article, error) {
			return ctx.PubMed.GetArticle(ctx.Identifier)
		}),
		IOE.Map[error](func(a *literature.Article) WithPubMedArticle {
			return WithPubMedArticle{
				WithPubMedClient: ctx,
				Article:          a,
			}
		}),
	)
}

func processPubMedFlow(
	ctx WithPubMedArticle,
) IOE.IOEither[error, any] {
	return F.Pipe4(
		IOE.Of[error](
			DownloadContext{
				WithPubMedClient: ctx.WithPubMedClient,
				PMID:             ctx.Article.PMID,
			},
		),
		IOE.Chain(checkPubMedAvailability),
		IOE.Map[error](setTargetFilename),
		IOE.Chain(downloadFromPubMed),
		IOE.Map[error](
			F.Constant1[DownloadContext, any](nil),
		),
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
