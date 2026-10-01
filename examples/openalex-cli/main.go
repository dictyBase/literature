package main

import (
	"context"
	"log"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/literature"
	"github.com/urfave/cli/v3"
)

// argsUsage is the shared identifier placeholder for subcommands.
const argsUsage = "<OpenAlexID|PMID|DOI>"

func main() {
	cmd := &cli.Command{
		Name:  "openalex-cli",
		Usage: "Extract publication data from the OpenAlex API",
		Description: "Fetch scholarly work metadata, citation impact metrics,\n" +
			"referenced works, and citing works from OpenAlex.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "email",
				Usage: "Contact email for the OpenAlex Polite Pool",
			},
			&cli.StringFlag{
				Name:  "api-key",
				Usage: "OpenAlex Premium API key (redacted from errors)",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: 5,
				Usage: "Page size for the citing command (max 200)",
			},
			&cli.StringFlag{
				Name:  "sort",
				Usage: "Sort expression, e.g. cited_by_count:desc",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "work",
				Usage:     "Fetch work metadata",
				ArgsUsage: argsUsage,
				Action:    runWork,
			},
			{
				Name:      "metrics",
				Usage:     "Fetch citation count, FWCI, and percentiles",
				ArgsUsage: argsUsage,
				Action:    runMetrics,
			},
			{
				Name:      "refs",
				Usage:     "List all works cited by the article",
				ArgsUsage: argsUsage,
				Action:    runRefs,
			},
			{
				Name:      "citing",
				Usage:     "One page of works citing the article",
				ArgsUsage: argsUsage,
				Action:    runCiting,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// newLogger builds the stderr logger shared by all actions.
func newLogger() *log.Logger {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}

// seedState validates the identifier and builds the initial state.
func seedState(input ActionInput) IOE.IOEither[error, State] {
	return F.Pipe2(
		State{
			Identifier: input.Identifier,
			Email:      input.Email,
			APIKey:     input.APIKey,
			Limit:      input.Limit,
			Sort:       input.Sort,
			Logger:     input.Logger,
		},
		E.FromPredicate(
			hasIdentifier,
			func(State) error {
				return cli.Exit(
					"Please provide an identifier (OpenAlex ID, PMID, or DOI)",
					1,
				)
			},
		),
		IOE.FromEither[error, State],
	)
}

// runWork fetches and logs work metadata for the identifier.
func runWork(_ context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		ActionInput{
			Identifier: cmd.Args().First(),
			Email:      cmd.String("email"),
			APIKey:     cmd.String("api-key"),
			Limit:      cmd.Int("limit"),
			Sort:       cmd.String("sort"),
			Logger:     newLogger(),
		},
		seedState,
		IOE.Bind(clientLens.Set, createOpenAlexClient),
		IOE.Chain(fetchByShape),
		IOE.ChainFirstIOK[error](logWorkHeader),
		ioeitherutils.ToEither[error, State],
		E.ToError[State],
	)
}

// runMetrics fetches and logs citation impact metrics.
func runMetrics(_ context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		ActionInput{
			Identifier: cmd.Args().First(),
			Email:      cmd.String("email"),
			APIKey:     cmd.String("api-key"),
			Limit:      cmd.Int("limit"),
			Sort:       cmd.String("sort"),
			Logger:     newLogger(),
		},
		seedState,
		IOE.Bind(clientLens.Set, createOpenAlexClient),
		IOE.Chain(bindMetrics),
		IOE.ChainFirstIOK[error](logMetrics),
		ioeitherutils.ToEither[error, State],
		E.ToError[State],
	)
}

// runRefs fetches and logs all works cited by the article.
func runRefs(_ context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		ActionInput{
			Identifier: cmd.Args().First(),
			Email:      cmd.String("email"),
			APIKey:     cmd.String("api-key"),
			Limit:      cmd.Int("limit"),
			Sort:       cmd.String("sort"),
			Logger:     newLogger(),
		},
		seedState,
		IOE.Bind(clientLens.Set, createOpenAlexClient),
		IOE.Chain(bindRefs),
		IOE.ChainFirstIOK[error](logRefs),
		ioeitherutils.ToEither[error, State],
		E.ToError[State],
	)
}

// runCiting fetches and logs one page of works citing the article.
func runCiting(_ context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		ActionInput{
			Identifier: cmd.Args().First(),
			Email:      cmd.String("email"),
			APIKey:     cmd.String("api-key"),
			Limit:      cmd.Int("limit"),
			Sort:       cmd.String("sort"),
			Logger:     newLogger(),
		},
		seedState,
		IOE.Bind(clientLens.Set, createOpenAlexClient),
		IOE.Chain(bindCiting),
		IOE.ChainFirstIOK[error](logCiting),
		ioeitherutils.ToEither[error, State],
		E.ToError[State],
	)
}

// bindMetrics fetches citation metrics and stores them on the state.
func bindMetrics(st State) IOE.IOEither[error, State] {
	return F.Pipe2(
		st,
		metricsBy,
		IOE.Map[error](func(m *literature.CitationMetrics) State {
			return metricsLens.Set(m)(st)
		}),
	)
}

// bindRefs fetches referenced works and stores them on the state.
func bindRefs(st State) IOE.IOEither[error, State] {
	return F.Pipe2(
		st,
		referencedEffect,
		IOE.Map[error](func(refs []*literature.OpenAlexWork) State {
			return refsLens.Set(refs)(st)
		}),
	)
}

// bindCiting fetches citing works and stores them on the state.
func bindCiting(st State) IOE.IOEither[error, State] {
	return F.Pipe2(
		st,
		citingWorksBy,
		IOE.Map[error](func(r *literature.OpenAlexWorksResult) State {
			return citingLens.Set(r)(st)
		}),
	)
}
