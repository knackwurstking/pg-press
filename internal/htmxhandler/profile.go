package htmxhandler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Profile struct {
	*Base
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/cookies", h.handleGetCookies)
	e.DELETE(h.ServerPathPrefix+"/cookies", h.handleDeleteCookies)
}

func (h *Profile) handleGetCookies(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	cookies, err := h.DB.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to list cookies: "+err.Error())
	}

	cookiesTable := components.CookiesTable(database.SortCookies(cookies))
	if err := cookiesTable.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cookies table: "+err.Error())
	}
	return nil
}

func (h *Profile) handleDeleteCookies(c echo.Context) error {
	value := utils.SanitizeInput(c.QueryParam("value"))
	if value == "" {
		return echo.NewHTTPError(http.StatusBadRequest,
			"cookie value parameter is required")
	}

	if err := h.DB.Cookies.Remove(value); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to delete cookie: "+err.Error())
	}

	return h.handleGetCookies(c)
}
