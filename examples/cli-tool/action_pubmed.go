package main

import (
	"errors"
	"fmt"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

var errPubMedPDFUnavailable = errors.New("PDF not available in PubMed")

func fetchPubMedArticle(state State) IOE.IOEither[error, State] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.Article, error) {
			return state.PubMed.GetArticle(state.Identifier)
		}),
		IOE.Map[error](func(a *literature.Article) State {
			return pubMedArticleLens.Set(a)(state)
		}),
	)
}

func processPubMedFlow(st State) IOE.IOEither[error, State] {
	return F.Pipe1(
		pmidLens.Set(st.PubMedArticle.PMID)(st),
		pubMedDownloadTail,
	)
}

// pubMedDownloadTail is the shared PubMed availability-check + download
// pipeline, reused by fallbackEurope and processPubMedFlow.
func pubMedDownloadTail(st State) IOE.IOEither[error, State] {
	return F.Pipe3(
		IOE.Of[error](st),
		IOE.Chain(checkPubMedAvailability),
		IOE.Let[error](targetFileLens.Set, targetFilename),
		IOE.Chain(downloadFromPubMed),
	)
}

func checkPubMedAvailability(state State) IOE.IOEither[error, State] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return state.PubMed.HasPDF(state.PMID)
		}),
		IOE.FilterOrElse(
			F.Identity[bool],
			F.Constant1[bool](errPubMedPDFUnavailable),
		),
		IOE.Map[error](F.Constant1[bool](state)),
	)
}

func downloadFromPubMed(st State) IOE.IOEither[error, State] {
	return F.Pipe1(
		IOE.TryCatchError(func() (State, error) {
			return st, st.PubMed.DownloadPDF(st.PMID, st.TargetFile)
		}),
		IOE.MapLeft[State](func(err error) error {
			return fmt.Errorf("failed to download PDF from PubMed: %w", err)
		}),
	)
}
