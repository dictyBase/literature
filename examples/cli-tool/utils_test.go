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
		t.Run(testCase.identifier, func(t *testing.T) {
			t.Parallel()
			st := State{Identifier: testCase.identifier}
			assert.Equal(t, testCase.expected, isDOI(st))
		})
	}
}

func TestHasIdentifier(t *testing.T) {
	t.Parallel()
	assert.True(t, hasIdentifier(State{Identifier: "12345678"}))
	assert.False(t, hasIdentifier(State{}))
}

func TestTargetFilename(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		state    State
		expected string
	}{
		{"custom output", State{PMID: "12345678", OutputFile: "paper.pdf"}, "paper.pdf"},
		{"default from PMID", State{PMID: "12345678"}, "12345678.pdf"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, testCase.expected, targetFilename(testCase.state))
		})
	}
}
