package html

import (
	"github.com/knackwurstking/pgpress/internal/web/html/auth"
	"github.com/knackwurstking/pgpress/internal/web/html/feed"
	"github.com/knackwurstking/pgpress/internal/web/html/home"
	"github.com/knackwurstking/pgpress/internal/web/html/profile"
	"github.com/knackwurstking/pgpress/internal/web/html/tools"
	"github.com/knackwurstking/pgpress/internal/web/html/troublereports"
)

type (
	Auth           = auth.Auth
	Feed           = feed.Feed
	Home           = home.Home
	Profile        = profile.Profile
	Tools          = tools.Tools
	TroubleReports = troublereports.TroubleReports
)

var (
	NewAuth           = auth.NewAuth
	NewFeed           = feed.NewFeed
	NewHome           = home.NewHome
	NewProfile        = profile.NewProfile
	NewTools          = tools.NewTools
	NewTroubleReports = troublereports.NewTroubleReports
)
