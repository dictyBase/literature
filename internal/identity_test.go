package internal

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityApply(t *testing.T) {
	t.Parallel()
	const (
		allFieldsAPIKey = "secret-key"
		onlyAPIKey      = "only-key"
	)
	testCases := []struct {
		name     string
		identity Identity
		want     url.Values
	}{
		{
			name: "all values",
			identity: Identity{
				APIKey: allFieldsAPIKey,
				Tool:   "literature-client",
				Email:  "dev@example.org",
			},
			want: url.Values{
				"api_key": {allFieldsAPIKey},
				"tool":    {"literature-client"},
				"email":   {"dev@example.org"},
			},
		},
		{
			name:     "api key only",
			identity: Identity{APIKey: onlyAPIKey},
			want: url.Values{
				"api_key": {onlyAPIKey},
			},
		},
		{
			name:     "none",
			identity: Identity{},
			want:     url.Values{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			values := url.Values{}

			testCase.identity.Apply(values)

			require.Equal(t, testCase.want, values)
		})
	}
}
