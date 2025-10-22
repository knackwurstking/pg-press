package utils

import (
	"github.com/a-h/templ"
)

func HXGetFeedList() templ.SafeURL {
	return buildURL("/htmx/feed/list", nil)
}
