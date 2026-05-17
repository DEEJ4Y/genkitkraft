package genkitchatprovider

import "net/url"

// sanitizeURL parses a raw URL and re-encodes its path segments and query
// parameters so that spaces and other unsafe characters are properly
// percent-encoded. Already-encoded sequences (e.g. %20) are preserved
// without double-encoding because Go's url.Parse decodes them first and
// url.Values.Encode / EscapedPath re-encodes them.
func sanitizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Re-encode path: EscapedPath() returns the properly encoded path.
	u.RawPath = u.EscapedPath()

	// Re-encode query parameters individually.
	if u.RawQuery != "" {
		q := u.Query() // parses and decodes
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}
