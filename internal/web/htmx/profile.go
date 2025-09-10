package htmx

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/profilepage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type Profile struct {
	DB *database.DB
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/profile/cookies", h.handleGetCookies),
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/profile/cookies", h.handleDeleteCookies),
		},
	)
}

func (h *Profile) handleGetCookies(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerProfile().Debug("Fetching cookies for user %s", user.Name)

	cookies, err := h.DB.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		logger.HTMXHandlerProfile().Error("Failed to list cookies for user %s: %v", user.Name, err)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to list cookies: "+err.Error())
	}

	logger.HTMXHandlerProfile().Debug("Found %d cookies for user %s", len(cookies), user.Name)

	cookiesTable := profilepage.Cookies(models.SortCookies(cookies))
	if err := cookiesTable.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerProfile().Error("Failed to render cookies table: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cookies table: "+err.Error())
	}
	return nil
}

func (h *Profile) handleDeleteCookies(c echo.Context) error {
	value := helpers.SanitizeInput(c.QueryParam("value"))
	if value == "" {
		logger.HTMXHandlerProfile().Error("Cookie deletion attempted with empty value")
		return echo.NewHTTPError(http.StatusBadRequest,
			"cookie value parameter is required")
	}

	logger.HTMXHandlerProfile().Info("Deleting cookie with value: %s", value)

	if err := h.DB.Cookies.Remove(value); err != nil {
		logger.HTMXHandlerProfile().Error("Failed to delete cookie: %v", err)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to delete cookie: "+err.Error())
	}

	logger.HTMXHandlerProfile().Info("Successfully deleted cookie")
	return h.handleGetCookies(c)
}
