package main

import (
	"fmt"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictybase/literature"
)

// createOpenAlexClient builds the client from the state's credential
// fields. It is a Kleisli arrow for IOE.Bind with clientLens.Set.
func createOpenAlexClient(
	st State,
) IOE.IOEither[error, *literature.OpenAlexClient] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.OpenAlexClient, error) {
			return literature.NewOpenAlexClient(clientOptions(st)...)
		}),
		IOE.MapLeft[*literature.OpenAlexClient](func(err error) error {
			return fmt.Errorf("create OpenAlex client: %w", err)
		}),
	)
}

// clientOptions is the pure state-to-options projection.
func clientOptions(st State) []literature.OpenAlexOption {
	return []literature.OpenAlexOption{
		literature.WithOpenAlexEmail(st.Email),
		literature.WithOpenAlexAPIKey(st.APIKey),
	}
}
