package html

import (
	"github.com/knackwurstking/pgpress/internal/web/html/profile"
	"github.com/knackwurstking/pgpress/internal/web/html/tools"
	"github.com/knackwurstking/pgpress/internal/web/html/troublereports"
)

type (
	Profile        = profile.Profile
	Tools          = tools.Tools
	TroubleReports = troublereports.TroubleReports
)

var (
	NewProfile        = profile.NewProfile
	NewTools          = tools.NewTools
	NewTroubleReports = troublereports.NewTroubleReports
)
