package handlers

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/logger"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Profile struct {
	*Base
}

func NewProfile(db *services.Registry) *Profile {
	return &Profile{
		Base: NewBase(db, logger.NewComponentLogger("Profile")),
	}
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
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

func (h *Profile) GetProfilePage(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	h.Log.Debug("Rendering profile page for user %s", user.Name)

	if err = h.handleUserNameChange(c, user); err != nil {
		return HandleError(err, "error updating username")
	}

	page := components.PageProfile(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render profile page")
	}

	return nil
}

func (h *Profile) HTMXGetCookies(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	h.Log.Debug("Fetching cookies for user %s", user.Name)

	cookies, err := h.Registry.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return HandleError(err, "failed to list cookies")
	}

	h.Log.Debug("Found %d cookies for user %s", len(cookies), user.Name)

	cookiesTable := components.CookiesDetails(models.SortCookies(cookies))
	err = cookiesTable.Render(c.Request().Context(), c.Response())
	if err != nil {
		return HandleError(err, "failed to render cookies table")
	}

	return nil
}

func (h *Profile) HTMXDeleteCookies(c echo.Context) error {
	value, err := ParseQueryString(c, "value")
	if err != nil {
		return HandleBadRequest(err, "failed to parse value")
	}

	h.Log.Info("Deleting cookie with value: %s", value)

	if err := h.Registry.Cookies.Remove(value); err != nil {
		return HandleError(err, "failed to delete cookie")
	}

	return h.HTMXGetCookies(c)
}

func (h *Profile) handleUserNameChange(c echo.Context, user *models.User) error {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < env.UserNameMinLength || len(userName) > env.UserNameMaxLength {
		return errors.NewValidationError(
			"user-name: username must be between 1 and 100 characters")
	}

	h.Log.Info("User %s (Telegram ID: %d) is changing username to %s",
		user.Name, user.TelegramID, userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.Registry.Users.Update(updatedUser); err != nil {
		return err
	}

	// Create feed entry
	feedTitle := "Benutzername geändert"
	feedContent := fmt.Sprintf("Alter Name: %s\nNeuer Name: %s", user.Name, userName)
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for username change: %v", err)
	}

	user.Name = userName

	return nil
}
