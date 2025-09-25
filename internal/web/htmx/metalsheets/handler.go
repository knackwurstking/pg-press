package metalsheets

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/dialogs"

	"github.com/labstack/echo/v4"
)

type MetalSheets struct {
	*handlers.BaseHandler
}

// TODO: Do not forget the feeds
func NewMetalSheets(db *database.DB) *MetalSheets {
	return &MetalSheets{
		BaseHandler: handlers.NewBaseHandler(db, logger.HTMXHandlerMetalSheets()),
	}
}

func (h *MetalSheets) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/metal-sheets/edit",
				h.GetEditDialog),
		},
	)
}

func (h *MetalSheets) GetEditDialog(c echo.Context) error {
	renderProps := &dialogs.EditMetalSheetProps{}

	var (
		toolID int64
		err    error
	)

	// Open edit dialog for adding or editing a metal sheet entry
	// First get the metal sheet id from query param
	if metalSheetID, _ := h.ParseInt64Query(c, "id"); metalSheetID > 0 {
		// Render dialog content for editing an existing metal sheet
		// Store metal sheet to render props
		if renderProps.MetalSheet, err = h.DB.MetalSheets.Get(metalSheetID); err != nil {
			return h.HandleError(c, err, "failed to fetch metal sheet from database")
		}
		toolID = renderProps.MetalSheet.ToolID
	} else {
		// No ID, render dialog content for adding a new metal sheet
		if toolID, err = h.ParseInt64Query(c, "tool_id"); err != nil {
			return h.HandleError(c, err, "failed to get the tool id from query")
		}
	}

	// Store tool to render props
	if renderProps.Tool, err = h.DB.Tools.Get(toolID); err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	d := dialogs.EditMetalSheet(renderProps)
	if err := d.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render edit metal sheet dialog: "+err.Error())
	}

	return nil
}
