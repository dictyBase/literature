package main

import (
	"fmt"
	"net/http"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	IOEF "github.com/IBM/fp-go/v2/ioeither/file"
	IOEH "github.com/IBM/fp-go/v2/ioeither/http"
)

func getFilename(pmid, customName string) string {
	if customName != "" {
		return customName
	}
	return fmt.Sprintf("%s.pdf", pmid)
}

func setTargetFilename(ctx DownloadContext) DownloadContext {
	ctx.TargetFile = getFilename(ctx.PMID, ctx.OutputFile)
	return ctx
}

func downloadPDF(ctx DownloadContext) IOE.IOEither[error, DownloadContext] {
	return F.Pipe3(
		ctx.PDFURL,
		IOEH.MakeGetRequest,
		IOEH.ReadAll(F.Pipe1(http.DefaultClient, IOEH.MakeClient)),
		IOE.Chain(func(data []byte) IOE.IOEither[error, DownloadContext] {
			return F.Pipe1(
				IOEF.WriteFile(ctx.TargetFile, 0o644)(data),
				IOE.Map[error](F.Constant1[[]byte](ctx)),
			)
		}),
	)
}
