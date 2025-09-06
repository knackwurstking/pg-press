package htmx

import (
	"net/http"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/services/cookie"
	"github.com/knackwurstking/pgpress/internal/logger"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	profilecomp "github.com/knackwurstking/pgpress/internal/web/templates/components/profile"

	"github.com/labstack/echo/v4"
)

type Profile struct {
	DB *database.DB
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/profile/cookies", h.handleGetCookies),
			webhelpers.NewEchoRoute(http.MethodDelete, "/htmx/profile/cookies", h.handleDeleteCookies),
		},
	)
}

func (h *Profile) handleGetCookies(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerProfile().Debug("Fetching cookies for user %s", user.UserName)

	cookies, err := h.DB.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		logger.HTMXHandlerProfile().Error("Failed to list cookies for user %s: %v", user.UserName, err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to list cookies: "+err.Error())
	}

	logger.HTMXHandlerProfile().Debug("Found %d cookies for user %s", len(cookies), user.UserName)

	cookiesTable := profilecomp.Cookies(cookie.SortCookies(cookies))
	if err := cookiesTable.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerProfile().Error("Failed to render cookies table: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cookies table: "+err.Error())
	}
	return nil
}

func (h *Profile) handleDeleteCookies(c echo.Context) error {
	value := webhelpers.SanitizeInput(c.QueryParam("value"))
	if value == "" {
		logger.HTMXHandlerProfile().Error("Cookie deletion attempted with empty value")
		return echo.NewHTTPError(http.StatusBadRequest,
			"cookie value parameter is required")
	}

	logger.HTMXHandlerProfile().Info("Deleting cookie with value: %s", value)

	if err := h.DB.Cookies.Remove(value); err != nil {
		logger.HTMXHandlerProfile().Error("Failed to delete cookie: %v", err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to delete cookie: "+err.Error())
	}

	logger.HTMXHandlerProfile().Info("Successfully deleted cookie")
	return h.handleGetCookies(c)
}
