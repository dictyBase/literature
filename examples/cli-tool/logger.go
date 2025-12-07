package main

import (
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
)

var (
	pubClientLogger = F.Curry2(
		func(msg string, ctx WithPubMedClient) IO.IO[WithPubMedClient] {
			return func() WithPubMedClient {
				ctx.Logger.Print(msg)
				return ctx
			}
		},
	)

	logEuropeArticle = func(ctx WithEuropePMCArticle) IO.IO[WithEuropePMCArticle] {
		return func() WithEuropePMCArticle {
			ctx.Logger.Println(
				"Article Details (EuropePMC)",
				"title", ctx.Article.Title,
				"authors", ctx.Article.AuthorString,
				"pmid", ctx.Article.PMID,
				"doi", ctx.Article.DOI,
			)
			return ctx
		}
	}

	logPubMedArticle = func(ctx WithPubMedArticle) IO.IO[WithPubMedArticle] {
		return func() WithPubMedArticle {
			ctx.Logger.Println(
				"Article Details (PubMed)",
				"title", ctx.Article.Title,
				"pmid", ctx.Article.PMID,
				"doi", ctx.Article.DOI,
			)
			return ctx
		}
	}
)
