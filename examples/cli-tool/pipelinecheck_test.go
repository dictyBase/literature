package main

import (
	"testing"

	"github.com/dictyBase/fp-go-loom/pipelinecheck"
)

func TestPipelineContinuity(t *testing.T) {
	t.Parallel()
	pipelinecheck.Require(t, pipelinecheck.Config{
		Roots:                []string{"."},
		RequirePointFreeSeed: true,
	})
}
