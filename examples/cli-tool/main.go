package main

import (
	"log"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "lit-cli",
		Usage:     "Fetch article metadata and download PDF (EuropePMC with PubMed fallback)",
		ArgsUsage: "<PMID|DOI>",
		Description: "Fetch article metadata for a PubMed ID or DOI and download the\n" +
			"PDF. Tries EuropePMC first, then falls back to PubMed.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output filename for PDF (default: <PMID>.pdf)",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	logger := log.Default()
	logger.SetOutput(os.Stderr)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return F.Pipe6(
		ActionInput{
			Identifier: ctx.Args().First(),
			OutputFile: ctx.String("output"),
			Logger:     logger,
		},
		seedState,
		IOE.Bind(europeLens.Set, createEuropeClient),
		IOE.Bind(pubMedLens.Set, createPubMedClient),
		IOE.Chain(fetchAndDownload),
		ioeitherutils.ToEither[error, State],
		E.ToError[State],
	)
}

func seedState(in ActionInput) IOE.IOEither[error, State] {
	return F.Pipe2(
		State{
			Identifier: in.Identifier,
			OutputFile: in.OutputFile,
			Logger:     in.Logger,
		},
		E.FromPredicate(
			hasIdentifier,
			func(State) error { return cli.Exit("Please provide a PMID or DOI", 1) },
		),
		IOE.FromEither[error, State],
	)
}
