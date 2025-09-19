package htmx

import (
	"github.com/knackwurstking/pgpress/internal/web/htmx/cycles"
	"github.com/knackwurstking/pgpress/internal/web/htmx/feed"
	"github.com/knackwurstking/pgpress/internal/web/htmx/profile"
	"github.com/knackwurstking/pgpress/internal/web/htmx/tools"
)

type (
	Feed    = feed.Feed
	Profile = profile.Profile
	Tools   = tools.Tools
	Cycles  = cycles.Cycles
)

var (
	NewFeed    = feed.NewFeed
	NewProfile = profile.NewProfile
	NewTools   = tools.NewTools
	NewCycles  = cycles.NewCycles
)
