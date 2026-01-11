package urlb

import (
	"fmt"
	"net/url"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/env"
)

// BuildURL constructs a URL with the given path and query parameters
func BuildURL(path string) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("%s%s", env.ServerPathPrefix, path))
}

// BuildURLWithParams constructs a URL with the given path and query parameters
func BuildURLWithParams(path string, params map[string]string) templ.SafeURL {
	values := url.Values{}
	for k, v := range params {
		if v == "" {
			continue
		}
		values.Add(k, v)
	}
	if len(values) > 0 {
		return BuildURL(fmt.Sprintf("%s?%s", path, values.Encode()))
	}
	return BuildURL(path)
}
