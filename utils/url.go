package utils

import (
	"fmt"
	"net/url"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
)

// BuildURL constructs a URL with the given path and query parameters
func BuildURL(path string, params map[string]string) templ.SafeURL {
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

func URLPressRegenerationsPage(press models.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tools/press/%d/regenerations", press), nil)
}
