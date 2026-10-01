package main

import (
	"fmt"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/predicate"
	"github.com/dictybase/literature"
)

// workByPMID fetches the work through the PMID-specific method.
func workByPMID(st State) IOE.IOEither[error, State] {
	return workBy((*literature.OpenAlexClient).GetWorkByPMID)(st)
}

// workGeneric fetches the work through the generic GetWork method.
func workGeneric(st State) IOE.IOEither[error, State] {
	return workBy((*literature.OpenAlexClient).GetWork)(st)
}

// workBy lifts a client fetch method over the state, storing the work.
func workBy(
	fetch func(*literature.OpenAlexClient, string) (*literature.OpenAlexWork, error),
) func(State) IOE.IOEither[error, State] {
	return func(state State) IOE.IOEither[error, State] {
		return F.Pipe1(
			workEffect(fetch)(state),
			IOE.Map[error](func(w *literature.OpenAlexWork) State {
				return workLens.Set(w)(state)
			}),
		)
	}
}

// workEffect performs the raw fetch, wrapping only in MapLeft.
func workEffect(
	fetch func(*literature.OpenAlexClient, string) (*literature.OpenAlexWork, error),
) func(State) IOE.IOEither[error, *literature.OpenAlexWork] {
	return func(state State) IOE.IOEither[error, *literature.OpenAlexWork] {
		return F.Pipe1(
			IOE.TryCatchError(func() (*literature.OpenAlexWork, error) {
				return fetch(state.Client, identifierLens.Get(state))
			}),
			IOE.MapLeft[*literature.OpenAlexWork](func(err error) error {
				return fmt.Errorf(
					"fetch work %s: %w",
					identifierLens.Get(state),
					err,
				)
			}),
		)
	}
}

// fetchByShape dispatches the PMID arm against the generic arm; applied
// to the state it fetches via GetWorkByPMID for PMID-shaped identifiers
// and via GetWork otherwise.
var fetchByShape = P.Fold(workGeneric, workByPMID)(isPMIDIdentifier)

// referencedEffect performs the raw GetReferencedWorks call.
func referencedEffect(state State) IOE.IOEither[error, []*literature.OpenAlexWork] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([]*literature.OpenAlexWork, error) {
			return state.Client.GetReferencedWorks(
				identifierLens.Get(state),
			)
		}),
		IOE.MapLeft[[]*literature.OpenAlexWork](func(err error) error {
			return fmt.Errorf(
				"fetch referenced works %s: %w",
				identifierLens.Get(state),
				err,
			)
		}),
	)
}

// citingWorksBy fetches one page of citing works with state flags.
func citingWorksBy(state State) IOE.IOEither[error, *literature.OpenAlexWorksResult] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.OpenAlexWorksResult, error) {
			return state.Client.GetCitingWorks(
				identifierLens.Get(state),
				literature.WithOpenAlexPerPage(state.Limit),
				literature.WithOpenAlexSort(state.Sort),
			)
		}),
		IOE.MapLeft[*literature.OpenAlexWorksResult](func(err error) error {
			return fmt.Errorf(
				"fetch citing works %s: %w",
				identifierLens.Get(state),
				err,
			)
		}),
	)
}

// metricsBy fetches citation metrics for the state's identifier.
func metricsBy(state State) IOE.IOEither[error, *literature.CitationMetrics] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.CitationMetrics, error) {
			return state.Client.GetCitationMetrics(
				identifierLens.Get(state),
			)
		}),
		IOE.MapLeft[*literature.CitationMetrics](func(err error) error {
			return fmt.Errorf(
				"fetch citation metrics %s: %w",
				identifierLens.Get(state),
				err,
			)
		}),
	)
}
