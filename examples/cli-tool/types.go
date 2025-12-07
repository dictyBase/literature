package main

import (
	"log"

	"github.com/dictybase/literature"
)

// RunContext is the base context for the CLI tool.
type RunContext struct {
	Identifier string
	OutputFile string
	Logger     *log.Logger
}

// WithEuropeClient adds the EuropePMC client to the context.
type WithEuropeClient struct {
	RunContext
	Europe *literature.EuropePMCClient
}

// WithPubMedClient adds the PubMed client to the context.
type WithPubMedClient struct {
	WithEuropeClient
	PubMed *literature.Client
}

type WithPubMedArticle struct {
	WithPubMedClient
	Article *literature.Article
}

type WithEuropePMCArticle struct {
	WithPubMedClient
	Article *literature.EuropePMCArticle
}

type DownloadContext struct {
	WithPubMedClient
	PMID       string
	PDFURL     string
	TargetFile string
}
