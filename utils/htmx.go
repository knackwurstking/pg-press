package utils

import (
	"fmt"
	"net/url"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/env"
)

// buildURL constructs a URL with the given path and query parameters
func buildURL(path string, params map[string]string) templ.SafeURL {
	u := fmt.Sprintf("%s%s", env.ServerPathPrefix, path)

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		u = fmt.Sprintf("%s?%s", u, values.Encode())
	}

	return templ.SafeURL(u)
}
