package internal

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	redactionTestValue     = "never-log-this-value"
	redactionTestOperation = "Get"
)

func TestRedactAPIKey_RedactsValue(t *testing.T) {
	t.Parallel()
	original := &url.Error{
		Op:  redactionTestOperation,
		URL: "https://eutils.example/efetch.fcgi?db=pubmed&api_key=" + redactionTestValue,
		Err: errors.New("dial failed"),
	}

	redacted := redactAPIKey(original)
	var urlError *url.Error
	req := require.New(t)
	req.ErrorAs(redacted, &urlError)
	req.Equal(
		"https://eutils.example/efetch.fcgi?api_key=REDACTED&db=pubmed",
		urlError.URL,
	)
	req.NotContains(redacted.Error(), redactionTestValue)
}

func TestRedactAPIKey_PreservesURLWithoutKey(t *testing.T) {
	t.Parallel()
	original := &url.Error{
		Op:  redactionTestOperation,
		URL: "https://eutils.example/efetch.fcgi?db=pubmed",
		Err: errors.New("dial failed"),
	}

	require.Same(t, original, redactAPIKey(original))
}

func TestRedactAPIKey_PassesThroughNonURLError(t *testing.T) {
	t.Parallel()
	original := errors.New("plain failure")

	require.Same(t, original, redactAPIKey(original))
}

func TestRedactAPIKey_HidesUnparseableURL(t *testing.T) {
	t.Parallel()
	original := &url.Error{
		Op:  redactionTestOperation,
		URL: "http://%zz/?api_key=" + redactionTestValue,
		Err: errors.New("invalid URL"),
	}

	redacted := redactAPIKey(original)
	var urlError *url.Error
	req := require.New(t)
	req.ErrorAs(redacted, &urlError)
	req.Equal("[redacted]", urlError.URL)
	req.NotContains(redacted.Error(), redactionTestValue)
}
