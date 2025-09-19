package htmx

import (
	"github.com/knackwurstking/pgpress/internal/web/htmx/feed"
	"github.com/knackwurstking/pgpress/internal/web/htmx/profile"
	"github.com/knackwurstking/pgpress/internal/web/htmx/tools"
)

type (
	Feed    = feed.Feed
	Profile = profile.Profile
	Tools   = tools.Tools
)

var (
	NewFeed    = feed.NewFeed
	NewProfile = profile.NewProfile
	NewTools   = tools.NewTools
)
