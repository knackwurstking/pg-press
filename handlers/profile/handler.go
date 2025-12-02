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
	"github.com/knackwurstking/ui"
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

	slog.Info("Rendering profile page", "user_name", user.Name)

	if err := h.handleUserNameChange(c, user); err != nil {
		return errors.Handler(err, "error updating username")
	}

	page := templates.Page(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render profile page")
	}

	return nil
}

func (h *Handler) HTMXGetCookies(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	slog.Info("Fetching cookies for user", "user_name", user.Name)

	cookies, err := h.registry.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return errors.Handler(err, "list cookies")
	}

	slog.Info("Found cookies for user", "cookies", len(cookies), "user_name", user.Name)

	cookiesTable := templates.Cookies(models.SortCookies(cookies))
	err = cookiesTable.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.Handler(err, "render cookies table")
	}

	return nil
}

func (h *Handler) HTMXDeleteCookies(c echo.Context) error {
	value, err := utils.ParseQueryString(c, "value")
	if err != nil {
		return errors.Handler(err, "parse value")
	}

	slog.Info("Deleting cookie", "value", utils.MaskString(value))

	if err := h.registry.Cookies.Remove(value); err != nil {
		return errors.Handler(err, "delete cookie")
	}

	return h.HTMXGetCookies(c)
}

func (h *Handler) handleUserNameChange(c echo.Context, user *models.User) error {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < env.UserNameMinLength || len(userName) > env.UserNameMaxLength {
		return errors.NewValidationError(
			"user-name: username must be between 1 and 100 characters")
	}

	slog.Info("Changing user name", "user_from", user.Name, "user_to", userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.registry.Users.Update(updatedUser); err != nil {
		return err
	}

	// Create feed entry
	feedTitle := "Benutzername geändert"
	feedContent := fmt.Sprintf("Alter Name: %s\nNeuer Name: %s", user.Name, userName)
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed for username change", "error", err)
	}

	user.Name = userName

	return nil
}
