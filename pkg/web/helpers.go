package web

import (
	"net/url"
)

// injectPayload injects a payload into a URL parameter
func injectPayload(targetURL, param, payload string) string {
	u, err := url.Parse(targetURL)
	if err != nil {
		return targetURL + "?" + param + "=" + url.QueryEscape(payload)
	}

	q := u.Query()
	q.Set(param, payload)
	u.RawQuery = q.Encode()
	return u.String()
}
