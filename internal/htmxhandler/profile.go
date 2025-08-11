package htmxhandler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
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
		return utils.HandlepgpressError(c, err)
	}

	// TODO: Continue here...
	return utils.HandleTemplate(c, database.SortCookies(cookies), h.Templates,
		[]string{constants.HTMXProfileCookiesTemplatePath})
}

func (h *Profile) handleDeleteCookies(c echo.Context) error {
	value := utils.SanitizeInput(c.QueryParam("value"))
	if value == "" {
		return echo.NewHTTPError(http.StatusBadRequest,
			"cookie value parameter is required")
	}

	if err := h.DB.Cookies.Remove(value); err != nil {
		return utils.HandlepgpressError(c, err)
	}

	return h.handleGetCookies(c)
}
