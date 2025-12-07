package main

import (
	"log"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	RIE "github.com/IBM/fp-go/v2/readerioeither"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "lit-cli",
		Usage: "Fetch article metadata and download PDF (EuropePMC with PubMed fallback)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output filename for PDF",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	identifier := ctx.Args().First()
	if S.IsEmpty(identifier) {
		return cli.Exit("Please provide a PMID or DOI", 1)
	}

	logger := log.Default()
	logger.SetOutput(os.Stderr)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Construct the program
	return F.Pipe5(
		IOE.Do[error](RunContext{
			Identifier: identifier,
			OutputFile: ctx.String("output"),
			Logger:     logger,
		}),
		IOE.Bind(InjectEuropeClient, createEuropeClient),
		IOE.Bind(InjectPubMedClient, createPubMedClient),
		IOE.Chain(
			RIE.MonadAlt(
				ExecuteEuropeFlow,
				func() RIE.ReaderIOEither[WithPubMedClient, error, any] {
					return ExecutePubMedFlow
				},
			),
		),
		ToEither,
		E.Fold(
			F.Identity[error],
			F.Constant1[any, error](nil),
		),
	)
}
