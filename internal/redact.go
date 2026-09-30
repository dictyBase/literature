package internal

import (
	"errors"
	"net/url"
)

func redactAPIKey(err error) error {
	var urlError *url.Error
	if !errors.As(err, &urlError) {
		return err
	}

	parsedURL, parseErr := url.Parse(urlError.URL)
	if parseErr != nil {
		return &url.Error{
			Op:  urlError.Op,
			URL: "[redacted]",
			Err: urlError.Err,
		}
	}

	query := parsedURL.Query()
	if query.Get("api_key") == "" {
		return err
	}
	query.Set("api_key", "REDACTED")
	parsedURL.RawQuery = query.Encode()

	return &url.Error{
		Op:  urlError.Op,
		URL: parsedURL.String(),
		Err: urlError.Err,
	}
}
