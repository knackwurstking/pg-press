package profile

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Profile struct {
	*handlers.BaseHandler
}

func NewProfile(db *database.DB) *Profile {
	return &Profile{
		BaseHandler: handlers.NewBaseHandler(db, logger.HTMXHandlerProfile()),
	}
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/profile/cookies",
				h.HandleGetCookies),
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/profile/cookies",
				h.HandleDeleteCookies),
		},
	)
}

func (h *Profile) HandleGetCookies(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("Fetching cookies for user %s", user.Name)

	cookies, err := h.DB.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return h.HandleError(c, err, "failed to list cookies: "+err.Error())
	}

	h.LogDebug("Found %d cookies for user %s", len(cookies), user.Name)

	cookiesTable := Cookies(models.SortCookies(cookies))
	err = cookiesTable.Render(c.Request().Context(), c.Response())
	if err != nil {
		return h.RenderInternalError(c,
			"failed to render cookies table: "+err.Error())
	}

	return nil
}

func (h *Profile) HandleDeleteCookies(c echo.Context) error {
	value, err := h.ParseStringQuery(c, "value")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	h.LogInfo("Deleting cookie with value: %s", value)

	if err := h.DB.Cookies.Remove(value); err != nil {
		return h.HandleError(c, err, "failed to delete cookie")
	}

	return h.HandleGetCookies(c)
}
