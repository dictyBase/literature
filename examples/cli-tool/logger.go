package main

import (
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
)

var (
	pubClientLogger = F.Curry2(
		func(msg string, st State) IO.IO[State] {
			return func() State {
				st.Logger.Print(msg)
				return st
			}
		},
	)

	logEuropeArticle = func(state State) IO.IO[State] {
		return func() State {
			state.Logger.Println(
				"Article Details (EuropePMC)",
				"title", state.EuropeArticle.Title,
				"authors", state.EuropeArticle.AuthorString,
				"pmid", state.EuropeArticle.PMID,
				"doi", state.EuropeArticle.DOI,
			)
			return state
		}
	}

	logPubMedArticle = func(state State) IO.IO[State] {
		return func() State {
			state.Logger.Println(
				"Article Details (PubMed)",
				"title", state.PubMedArticle.Title,
				"pmid", state.PubMedArticle.PMID,
				"doi", state.PubMedArticle.DOI,
			)
			return state
		}
	}
)
