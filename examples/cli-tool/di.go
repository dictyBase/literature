package main

import (
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

func createEuropeClient(
	_ State,
) IOE.IOEither[error, *literature.EuropePMCClient] {
	return IOE.TryCatchError(func() (*literature.EuropePMCClient, error) {
		return literature.NewEuropePMCClient()
	})
}

func createPubMedClient(
	_ State,
) IOE.IOEither[error, *literature.Client] {
	return IOE.TryCatchError(func() (*literature.Client, error) {
		return literature.New()
	})
}
