package main

import (
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

func ExecuteEuropeFlow(
	ctx WithPubMedClient,
) IOE.IOEither[error, any] {
	return F.Pipe3(
		IOE.Of[error](ctx),
		IOE.Chain(F.Ternary(isDOI, europeByDOI, europeByPMID)),
		IOE.ChainFirstIOK[error](logEuropeArticle),
		IOE.Chain(
			F.Ternary(
				hasEuropePDF,
				downloadEuropePDF,
				fallbackEurope,
			),
		),
	)
}

func ExecutePubMedFlow(
	ctx WithPubMedClient,
) IOE.IOEither[error, any] {
	return F.Pipe4(
		IOE.Of[error](ctx),
		IOE.ChainFirstIOK[error](
			pubClientLogger("Not found in EuropePMC. Trying PubMed..."),
		),
		IOE.Chain(fetchPubMedArticle),
		IOE.ChainFirstIOK[error](logPubMedArticle),
		IOE.Chain(processPubMedFlow),
	)
}
