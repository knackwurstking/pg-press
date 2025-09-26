package htmx

import (
	"github.com/knackwurstking/pgpress/internal/web/htmx/cycles"
	"github.com/knackwurstking/pgpress/internal/web/htmx/metalsheets"
	"github.com/knackwurstking/pgpress/internal/web/htmx/nav"
	"github.com/knackwurstking/pgpress/internal/web/htmx/tools"
	"github.com/knackwurstking/pgpress/internal/web/htmx/troublereports"
)

type (
	Nav            = nav.Nav
	Tools          = tools.Tools
	Cycles         = cycles.Cycles
	TroubleReports = troublereports.TroubleReports
	MetalSheets    = metalsheets.MetalSheets
)

var (
	NewNav            = nav.NewNav
	NewTools          = tools.NewTools
	NewCycles         = cycles.NewCycles
	NewTroubleReports = troublereports.NewTroubleReports
	NewMetalSheets    = metalsheets.NewMetalSheets
)
