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
	e.GET(serverPathPrefix+"/htmx/tools/edit", h.handleEdit)
}

func (h *Tools) handleListAll(c echo.Context) error {
	// TODO: Implement list all handler

	return errors.New("under construction")
}

// handleEdit renders a dialog for editing or creating a tool
func (h *Tools) handleEdit(c echo.Context) error {
	id, _ := utils.ParseInt64Query(c, constants.QueryParamID)
	close := utils.ParseBoolQuery(c, constants.QueryParamClose)

	toolEdit := components.ToolEditDialog(&components.ToolEditDialogProps{
		ID:    id,
		Close: close,
	})
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}
	return nil
}
