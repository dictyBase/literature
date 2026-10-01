package main

import (
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
)

// pmidPrefix is the identifier prefix OpenAlex uses for PubMed IDs.
const pmidPrefix = "PMID:"

var (
	// hasIdentifier guards that a subcommand received an argument.
	hasIdentifier = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(identifierLens.Get),
	)

	// isPMIDIdentifier reports whether the state's identifier is a bare
	// or PMID:-prefixed numeric identifier.
	isPMIDIdentifier = F.Pipe1(
		isDigitString,
		P.ContraMap(trimPMIDPrefix),
	)
)

// isDigitString reports whether the string is non-empty and all digits.
var isDigitString = func(s string) bool {
	return s != "" && strings.IndexFunc(s, nonDigit) < 0
}

// nonDigit is the rune predicate for non-digit characters.
func nonDigit(r rune) bool {
	return r < '0' || r > '9'
}

// trimPMIDPrefix strips a leading PMID: prefix from the state's
// identifier.
func trimPMIDPrefix(st State) string {
	raw := identifierLens.Get(st)
	return strings.TrimPrefix(raw, pmidPrefix)
}
