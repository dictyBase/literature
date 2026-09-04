package main

import (
	"fmt"
	"net/http"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	IOEF "github.com/IBM/fp-go/v2/ioeither/file"
	IOEH "github.com/IBM/fp-go/v2/ioeither/http"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
)

func targetFilename(st State) string {
	return F.Pipe1(
		st.OutputFile,
		P.Fold(
			F.Constant1[string](fmt.Sprintf("%s.pdf", st.PMID)),
			F.Identity[string],
		)(S.IsNonEmpty),
	)
}

func downloadPDF(state State) IOE.IOEither[error, State] {
	return F.Pipe3(
		state.PDFURL,
		IOEH.MakeGetRequest,
		IOEH.ReadAll(F.Pipe1(http.DefaultClient, IOEH.MakeClient)),
		IOE.Chain(func(data []byte) IOE.IOEither[error, State] {
			return F.Pipe1(
				IOEF.WriteFile(state.TargetFile, 0o644)(data),
				IOE.Map[error](F.Constant1[[]byte](state)),
			)
		}),
	)
}
