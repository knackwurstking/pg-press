package htmxhandler

import (
	"errors"
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	e.GET(serverPathPrefix+"/htmx/tools/list-all", h.handleListAll)

	e.GET(serverPathPrefix+"/htmx/tools/edit", func(c echo.Context) error {
		return h.handleEdit(nil, c)
	})
	e.POST(serverPathPrefix+"/htmx/tools/edit", h.handleEditPOST)
	e.PUT(serverPathPrefix+"/htmx/tools/edit", h.handleEditPUT)
}

func (h *Tools) handleListAll(c echo.Context) error {
	// TODO: Implement list all handler

	return errors.New("under construction")
}

// handleEdit renders a dialog for editing or creating a tool
func (h *Tools) handleEdit(props *components.ToolEditDialogProps, c echo.Context) error {
	if props == nil {
		props = &components.ToolEditDialogProps{}
	}

	props.ID, _ = utils.ParseInt64Query(c, constants.QueryParamID)
	props.Close = utils.ParseBoolQuery(c, constants.QueryParamClose)

	toolEdit := components.ToolEditDialog(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}
	return nil
}

func (h *Tools) handleEditPOST(c echo.Context) error {
	// TODO: Implement edit POST handler

	return errors.New("under construction")
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	// TODO: Implement edit PUT handler

	return errors.New("under construction")
}
