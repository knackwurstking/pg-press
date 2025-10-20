package helpers

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/env"
)

func HXGetFeedList() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/feed/list",
		env.ServerPathPrefix,
	))
}
