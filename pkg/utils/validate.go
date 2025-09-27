package utils

import (
	"net/url"
	"strings"
)

// IsValidCallbackURL returns true if the provided URL is a valid http(s) URL with a host.
// Empty string should be validated by callers as "not provided" rather than invalid.
func IsValidCallbackURL(u string) bool {
	if u == "" {
		return true
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	return true
}
