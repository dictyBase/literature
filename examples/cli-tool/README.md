# lit-cli

Example CLI for the `dictybase/literature` library. Fetches article metadata
and downloads the PDF. Tries EuropePMC first, then falls back to PubMed.

## Run

Build/run the **package**, not a single file:

```bash
# from examples/cli-tool/
go run . <PMID|DOI>

# custom output filename (default: <PMID>.pdf)
go run . -o paper.pdf <PMID|DOI>
```

```bash
# examples
go run . 12345678
go run . 10.1016/j.jmb.2010.08.037
go run . -o jmb-2010.pdf 10.1016/j.jmb.2010.08.037
```

## Usage

```
lit-cli [OPTIONS] <PMID|DOI>

OPTIONS:
  -o, --output FILE   Output filename for PDF  [default: <PMID>.pdf]
  -h, --help          Show help and exit

EXIT CODES:
  0   Success
  1   Missing identifier, article not found, or download failed
```

## Module setup

This example is its own Go module (`go.mod`) with
`replace github.com/dictybase/literature => ../../`, so it always builds
against the local checkout of the library.

Build commands that work:

```bash
go build .      # or
go build ./...  # or
go run .
```

`go build ./main.go` does **not** work — the package spans multiple files
(`types.go`, `di.go`, actions). Go compiles per package, not per file.
