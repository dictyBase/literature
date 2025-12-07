package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDOI(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		identifier string
		expected   bool
	}{
		{"10.1016/j.jmb.2010.08.037", true},
		{"12345678", false},
		{"PMC123456", false},
		{"doi/10.1234", true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.identifier, func(t *testing.T) {
			t.Parallel()
			ctx := WithPubMedClient{
				WithEuropeClient: WithEuropeClient{
					RunContext: RunContext{
						Identifier: testCase.identifier,
					},
				},
			}
			assert.Equal(t, testCase.expected, isDOI(ctx))
		})
	}
}
