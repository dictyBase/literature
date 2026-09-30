package internal

import "net/url"

// Identity carries the NCBI E-utilities identification parameters. All fields
// are optional; empty fields are omitted from the request.
type Identity struct {
	APIKey string
	Tool   string
	Email  string
}

// Apply adds the non-empty identification parameters to values.
func (identity Identity) Apply(values url.Values) {
	if identity.APIKey != "" {
		values.Set("api_key", identity.APIKey)
	}
	if identity.Tool != "" {
		values.Set("tool", identity.Tool)
	}
	if identity.Email != "" {
		values.Set("email", identity.Email)
	}
}
