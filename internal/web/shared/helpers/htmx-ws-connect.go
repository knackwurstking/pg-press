package helpers

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/env"
)

func HXWsConnectNavFeedCounter() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"connect:%s/htmx/nav/feed-counter",
		env.ServerPathPrefix,
	))
}
