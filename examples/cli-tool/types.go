package main

import (
	"log"

	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/literature"
)

// ActionInput is the boundary input for the CLI action.
type ActionInput struct {
	Identifier string
	OutputFile string
	Logger     *log.Logger
}

// State carries dependencies and intermediates through the pipeline.
type State struct {
	Identifier    string
	OutputFile    string
	Logger        *log.Logger
	Europe        *literature.EuropePMCClient
	PubMed        *literature.Client
	EuropeArticle *literature.EuropePMCArticle
	PubMedArticle *literature.Article
	PMID          string
	PDFURL        string
	TargetFile    string
}

var (
	identifierLens = L.MakeLens(
		func(s State) string { return s.Identifier },
		func(s State, v string) State { s.Identifier = v; return s },
	)
	europeLens = L.MakeLens(
		func(s State) *literature.EuropePMCClient { return s.Europe },
		func(s State, v *literature.EuropePMCClient) State { s.Europe = v; return s },
	)
	pubMedLens = L.MakeLens(
		func(s State) *literature.Client { return s.PubMed },
		func(s State, v *literature.Client) State { s.PubMed = v; return s },
	)
	europeArticleLens = L.MakeLens(
		func(s State) *literature.EuropePMCArticle { return s.EuropeArticle },
		func(s State, v *literature.EuropePMCArticle) State { s.EuropeArticle = v; return s },
	)
	pubMedArticleLens = L.MakeLens(
		func(s State) *literature.Article { return s.PubMedArticle },
		func(s State, v *literature.Article) State { s.PubMedArticle = v; return s },
	)
	pmidLens = L.MakeLens(
		func(s State) string { return s.PMID },
		func(s State, v string) State { s.PMID = v; return s },
	)
	pdfURLLens = L.MakeLens(
		func(s State) string { return s.PDFURL },
		func(s State, v string) State { s.PDFURL = v; return s },
	)
	targetFileLens = L.MakeLens(
		func(s State) string { return s.TargetFile },
		func(s State, v string) State { s.TargetFile = v; return s },
	)
)
