package main

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

func europeArticleBy(
	fetch func(*literature.EuropePMCClient, string) (*literature.EuropePMCArticle, error),
	state State,
) IOE.IOEither[error, State] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.EuropePMCArticle, error) {
			return fetch(state.Europe, state.Identifier)
		}),
		IOE.Map[error](func(a *literature.EuropePMCArticle) State {
			return europeArticleLens.Set(a)(state)
		}),
	)
}

func europeByDOI(st State) IOE.IOEither[error, State] {
	return europeArticleBy((*literature.EuropePMCClient).GetArticleByDOI, st)
}

func europeByPMID(st State) IOE.IOEither[error, State] {
	return europeArticleBy((*literature.EuropePMCClient).GetArticle, st)
}

func hasEuropePDF(st State) bool {
	return st.EuropeArticle.HasPDF
}

func getPDFURLs(state State) IOE.IOEither[error, State] {
	return F.Pipe3(
		IOE.TryCatchError(func() ([]literature.EuropePMCFullTextURL, error) {
			return state.Europe.GetPDFURLs(state.PMID)
		}),
		IOE.MapLeft[[]literature.EuropePMCFullTextURL](func(err error) error {
			return fmt.Errorf("fetch PDF URLs for PMID %s: %w", state.PMID, err)
		}),
		IOE.ChainEitherK(F.Flow2(
			A.Head,
			E.FromOption[literature.EuropePMCFullTextURL](func() error {
				return fmt.Errorf("no PDF URL listed for PMID %s", state.PMID)
			}),
		)),
		IOE.Map[error](func(u literature.EuropePMCFullTextURL) State {
			return pdfURLLens.Set(u.URL)(state)
		}),
	)
}

func downloadEuropePDF(st State) IOE.IOEither[error, State] {
	return F.Pipe3(
		IOE.Of[error](pmidLens.Set(st.EuropeArticle.PMID)(st)),
		IOE.Chain(getPDFURLs),
		IOE.Let[error](targetFileLens.Set, targetFilename),
		IOE.Chain(downloadPDF),
	)
}

func fallbackEurope(st State) IOE.IOEither[error, State] {
	return F.Pipe1(
		pmidLens.Set(st.EuropeArticle.PMID)(st),
		pubMedDownloadTail,
	)
}
