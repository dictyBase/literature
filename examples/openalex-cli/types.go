package main

import (
	"log"

	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/literature"
)

// ActionInput is the boundary input for each CLI subcommand action.
type ActionInput struct {
	Identifier string
	Email      string
	APIKey     string
	Limit      int
	Sort       string
	Logger     *log.Logger
}

// State carries dependencies and intermediates through the pipeline.
type State struct {
	Identifier string                          `json:"identifier"`
	Email      string                          `json:"-"`
	APIKey     string                          `json:"-"`
	Limit      int                             `json:"limit"`
	Sort       string                          `json:"sort,omitempty"`
	Logger     *log.Logger                     `json:"-"`
	Client     *literature.OpenAlexClient      `json:"-"`
	Work       *literature.OpenAlexWork        `json:"work,omitempty"`
	Metrics    *literature.CitationMetrics     `json:"metrics,omitempty"`
	Refs       []*literature.OpenAlexWork      `json:"refs,omitempty"`
	Citing     *literature.OpenAlexWorksResult `json:"citing,omitempty"`
}

var (
	identifierLens = L.MakeLens(
		func(s State) string { return s.Identifier },
		func(s State, v string) State { s.Identifier = v; return s },
	)
	clientLens = L.MakeLens(
		func(s State) *literature.OpenAlexClient { return s.Client },
		func(s State, v *literature.OpenAlexClient) State {
			s.Client = v
			return s
		},
	)
	workLens = L.MakeLens(
		func(s State) *literature.OpenAlexWork { return s.Work },
		func(s State, v *literature.OpenAlexWork) State { s.Work = v; return s },
	)
	metricsLens = L.MakeLens(
		func(s State) *literature.CitationMetrics { return s.Metrics },
		func(s State, v *literature.CitationMetrics) State {
			s.Metrics = v
			return s
		},
	)
	refsLens = L.MakeLens(
		func(s State) []*literature.OpenAlexWork { return s.Refs },
		func(s State, v []*literature.OpenAlexWork) State { s.Refs = v; return s },
	)
	citingLens = L.MakeLens(
		func(s State) *literature.OpenAlexWorksResult { return s.Citing },
		func(s State, v *literature.OpenAlexWorksResult) State {
			s.Citing = v
			return s
		},
	)
)
