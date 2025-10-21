package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/env"
)

func HXGetCookies() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/profile/cookies",
		env.ServerPathPrefix,
	))
}

func HXDeleteCookies(value string) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/profile/cookies?value=%s",
		env.ServerPathPrefix, value,
	))
}
