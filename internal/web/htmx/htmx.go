package htmx

import (
	"github.com/knackwurstking/pgpress/internal/web/htmx/metalsheets"
	"github.com/knackwurstking/pgpress/internal/web/htmx/nav"
)

type (
	Nav         = nav.Nav
	MetalSheets = metalsheets.MetalSheets
)

var (
	NewNav         = nav.NewNav
	NewMetalSheets = metalsheets.NewMetalSheets
)
