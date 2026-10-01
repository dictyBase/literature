# openalex-cli

FP-Go CLI example for extracting publication data from the
[OpenAlex API](https://openalex.org/) using the `literature` client.

Mirrors the pipeline structure of `examples/cli-tool`: point-free
`F.PipeN` entrypoints, a `State` struct with lenses threaded through
`IOE.Bind`/`Chain`, and loom predicates for identifier dispatch.

## Commands

```bash
go run . work    W2100837269                 # metadata (OpenAlex ID, PMID, or DOI)
go run . metrics 10.1038/nature12373         # citation count, FWCI, percentiles
go run . refs    PMID:23842501               # all works cited by the article
go run . citing  W2100837269 --limit 10     # one page of citing works
```

Global flags: `--email` (Polite Pool), `--api-key` (Premium, redacted
from errors), `--limit`, `--sort` (e.g. `cited_by_count:desc`).

## Layout

| File | Role |
|------|------|
| `main.go` | urfave/cli v3 commands + `run*` pipelines + `seedState` |
| `types.go` | `ActionInput`/`State` structs and lenses |
| `di.go` | `createOpenAlexClient` Kleisli arrow |
| `flows.go` | raw-effect fetchers + PMID dispatch via `P.Fold` |
| `logger.go` | `IO` taps rendering work details |
| `utils.go` | identifier predicates |
| `pipelinecheck_test.go` | enforces point-free seed and branching |

Pipeline shape per command:

```
ActionInput (boundary value)
  → seedState            validate identifier, build State
  → IOE.Bind(clientLens.Set, createOpenAlexClient)
  → IOE.Chain(fetch...)  raw effect in TryCatchError, wrap in MapLeft
  → IOE.ChainFirstIOK    render details to stderr
  → ioeitherutils.ToEither → E.ToError at the edge
```

Run `go test ./...` to execute the `pipelinecheck` style gates.
