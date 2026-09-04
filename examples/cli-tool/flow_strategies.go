package main

import (
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/predicate"
)

var (
	resolveEuropeFetch    = P.Fold(europeByPMID, europeByDOI)
	resolveEuropeDownload = P.Fold(fallbackEurope, downloadEuropePDF)
)

func executeEuropeFlow(st State) IOE.IOEither[error, State] {
	return F.Pipe3(
		IOE.Of[error](st),
		IOE.Chain(resolveEuropeFetch(isDOI)),
		IOE.ChainFirstIOK[error](logEuropeArticle),
		IOE.Chain(resolveEuropeDownload(hasEuropePDF)),
	)
}

func executePubMedFlow(st State) IOE.IOEither[error, State] {
	return F.Pipe4(
		IOE.Of[error](st),
		IOE.ChainFirstIOK[error](
			pubClientLogger("Not found in EuropePMC. Trying PubMed..."),
		),
		IOE.Chain(fetchPubMedArticle),
		IOE.ChainFirstIOK[error](logPubMedArticle),
		IOE.Chain(processPubMedFlow),
	)
}

func fetchAndDownload(st State) IOE.IOEither[error, State] {
	return F.Pipe1(
		executeEuropeFlow(st),
		IOE.Alt(func() IOE.IOEither[error, State] {
			return executePubMedFlow(st)
		}),
	)
}
