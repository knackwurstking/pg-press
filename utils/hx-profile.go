package utils

import "github.com/a-h/templ"

func HXGetCookies() templ.SafeURL {
	return BuildURL("/htmx/profile/cookies", nil)
}

func HXDeleteCookies(value string) templ.SafeURL {
	return BuildURL("/htmx/profile/cookies", map[string]string{
		"value": value,
	})
}
