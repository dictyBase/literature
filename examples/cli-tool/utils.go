package main

import (
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
)

func ToEither[A any](ioe IOE.IOEither[error, A]) E.Either[error, A] {
	return ioe()
}

func isDOI(ctx WithPubMedClient) bool {
	return F.Pipe1(
		ctx.Identifier,
		P.Or(S.Includes("/"))(S.HasPrefix("10.")),
	)
}
