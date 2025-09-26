package htmx

import (
	"github.com/knackwurstking/pgpress/internal/web/htmx/cycles"
	"github.com/knackwurstking/pgpress/internal/web/htmx/metalsheets"
	"github.com/knackwurstking/pgpress/internal/web/htmx/nav"
	"github.com/knackwurstking/pgpress/internal/web/htmx/troublereports"
)

type (
	Nav            = nav.Nav
	Cycles         = cycles.Cycles
	TroubleReports = troublereports.TroubleReports
	MetalSheets    = metalsheets.MetalSheets
)

var (
	NewNav            = nav.NewNav
	NewCycles         = cycles.NewCycles
	NewTroubleReports = troublereports.NewTroubleReports
	NewMetalSheets    = metalsheets.NewMetalSheets
)
