package utils

import (
	"github.com/a-h/templ"
)

func HXGetFeedList() templ.SafeURL {
	return BuildURL("/htmx/feed/list", nil)
}
