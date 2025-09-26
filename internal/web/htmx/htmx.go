package htmx

import (
	"github.com/knackwurstking/pgpress/internal/web/htmx/cycles"
	"github.com/knackwurstking/pgpress/internal/web/htmx/metalsheets"
	"github.com/knackwurstking/pgpress/internal/web/htmx/nav"
)

type (
	Nav         = nav.Nav
	Cycles      = cycles.Cycles
	MetalSheets = metalsheets.MetalSheets
)

var (
	NewNav         = nav.NewNav
	NewCycles      = cycles.NewCycles
	NewMetalSheets = metalsheets.NewMetalSheets
)
