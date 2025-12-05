package profile

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/profile/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			ui.NewEchoRoute(http.MethodGet, path, h.GetProfilePage),
			ui.NewEchoRoute(http.MethodGet, path+"/cookies", h.HTMXGetCookies),
			ui.NewEchoRoute(http.MethodDelete, path+"/cookies",
				h.HTMXDeleteCookies),
		},
	)
}

func (h *Handler) GetProfilePage(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	err := h.handleUserNameChange(c, user)
	if err != nil {
		return errors.HandlerError(err, "error updating username")
	}

	page := templates.Page(user)
	err = page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ProfilePage")
	}

	return nil
}

func (h *Handler) HTMXGetCookies(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	cookies, dberr := h.registry.Cookies.ListApiKey(user.ApiKey)
	if dberr != nil {
		return errors.HandlerError(dberr, "list cookies")
	}

	cookiesTable := templates.Cookies(models.SortCookies(cookies))
	err := cookiesTable.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cookies")
	}

	return nil
}

func (h *Handler) HTMXDeleteCookies(c echo.Context) error {
	slog.Info("Eat the cookie from a user :)")

	value, err := utils.ParseQueryString(c, "value")
	if err != nil {
		return errors.HandlerError(err, "parse value")
	}

	dberr := h.registry.Cookies.Remove(value)
	if dberr != nil {
		return errors.HandlerError(dberr, "delete cookie")
	}

	return h.HTMXGetCookies(c)
}

func (h *Handler) handleUserNameChange(c echo.Context, user *models.User) error {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < env.UserNameMinLength || len(userName) > env.UserNameMaxLength {
		return fmt.Errorf("user-name: username must be between 1 and 100 characters")
	}

	slog.Info("Changing user name", "user_from", user.Name, "user_to", userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	dberr := h.registry.Users.Update(updatedUser)
	if dberr != nil {
		return dberr
	}

	// Create feed entry
	feedTitle := "Benutzername geändert"
	feedContent := fmt.Sprintf("Alter Name: %s\nNeuer Name: %s", user.Name, userName)
	_, dberr = h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for username change", "error", dberr)
	}

	user.Name = userName

	return nil
}
