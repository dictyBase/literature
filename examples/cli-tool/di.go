package main

import (
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

// InjectEuropeClient sets the EuropePMC client in the context.
var InjectEuropeClient = F.Curry2(
	func(epc *literature.EuropePMCClient, ctx RunContext) WithEuropeClient {
		return WithEuropeClient{RunContext: ctx, Europe: epc}
	},
)

// InjectPubMedClient sets the PubMed client in the context.
var InjectPubMedClient = F.Curry2(
	func(epc *literature.Client, ctx WithEuropeClient) WithPubMedClient {
		return WithPubMedClient{WithEuropeClient: ctx, PubMed: epc}
	},
)

func createEuropeClient(
	_ RunContext,
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
