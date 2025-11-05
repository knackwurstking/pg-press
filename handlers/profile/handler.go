package profile

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/profile/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
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

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			utils.NewEchoRoute(http.MethodGet, "/profile", h.GetProfilePage),
			utils.NewEchoRoute(http.MethodGet, "/htmx/profile/cookies", h.HTMXGetCookies),
			utils.NewEchoRoute(http.MethodDelete, "/htmx/profile/cookies",
				h.HTMXDeleteCookies),
		},
	)
}

func (h *Handler) GetProfilePage(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	slog.Debug("Rendering profile page", "user_name", user.Name)

	if err = h.handleUserNameChange(c, user); err != nil {
		return utils.HandleError(err, "error updating username")
	}

	page := components.PageProfile(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render profile page")
	}

	return nil
}

func (h *Handler) HTMXGetCookies(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	slog.Debug("Fetching cookies for user", "user_name", user.Name)

	cookies, err := h.registry.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return utils.HandleError(err, "failed to list cookies")
	}

	slog.Debug("Found cookies for user", "cookies", len(cookies), "user_name", user.Name)

	cookiesTable := components.CookiesDetails(models.SortCookies(cookies))
	err = cookiesTable.Render(c.Request().Context(), c.Response())
	if err != nil {
		return utils.HandleError(err, "failed to render cookies table")
	}

	return nil
}

func (h *Handler) HTMXDeleteCookies(c echo.Context) error {
	value, err := utils.ParseQueryString(c, "value")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse value")
	}

	slog.Info("Deleting cookie", "value", utils.MaskString(value))

	if err := h.registry.Cookies.Remove(value); err != nil {
		return utils.HandleError(err, "failed to delete cookie")
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
