package urlb

import "github.com/a-h/templ"

func Profile() templ.SafeURL {
	return BuildURL("/profile")
}

func ProfileCookies(cookieValue string) templ.SafeURL {
	return BuildURLWithParams("/profile/cookies", map[string]string{
		"value": cookieValue,
	})
}
