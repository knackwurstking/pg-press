package urlb

import "github.com/a-h/templ"

func Home() templ.SafeURL {
	return BuildURL("/")
}
