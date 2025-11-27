package utils

import (
	"github.com/a-h/templ"
)

func HXWsConnectNavFeedCounter() templ.SafeURL {
	return BuildURL("/htmx/nav/feed-counter", nil)
}
