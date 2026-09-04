package main

import (
	F "github.com/IBM/fp-go/v2/function"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
)

var (
	isDOI = F.Pipe1(
		P.Or(S.Includes("/"))(S.HasPrefix("10.")),
		P.ContraMap(identifierLens.Get),
	)

	hasIdentifier = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(identifierLens.Get),
	)
)
