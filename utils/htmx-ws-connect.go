package utils

import (
	"github.com/a-h/templ"
)

func HXWsConnectNavFeedCounter() templ.SafeURL {
	return buildURL("/htmx/nav/feed-counter", nil)
}
