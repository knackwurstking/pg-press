package profile

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/features/profile/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler

	userNameMinLength int
	userNameMaxLength int
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler(db, logger.NewComponentLogger("Profile")),
		userNameMinLength: 1,
		userNameMaxLength: 100,
	}
}

func (h *Handler) GetProfilePage(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("Rendering profile page for user %s", user.Name)

	if err = h.handleUserNameChange(c, user); err != nil {
		return h.HandleError(c, err, "error updating username")
	}

	page := templates.ProfilePage(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render profile page: "+err.Error())
	}
	return nil
}

func (h *Handler) HTMXGetCookies(c echo.Context) error {
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

	cookiesTable := templates.CookiesDetails(models.SortCookies(cookies))
	err = cookiesTable.Render(c.Request().Context(), c.Response())
	if err != nil {
		return h.RenderInternalError(c,
			"failed to render cookies table: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXDeleteCookies(c echo.Context) error {
	value, err := h.ParseStringQuery(c, "value")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	h.LogInfo("Deleting cookie with value: %s", value)

	if err := h.DB.Cookies.Remove(value); err != nil {
		return h.HandleError(c, err, "failed to delete cookie")
	}

	return h.HTMXGetCookies(c)
}

func (h *Handler) handleUserNameChange(c echo.Context, user *models.User) error {
	userName := h.GetSanitizedFormValue(c, constants.UserNameFormField)

	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < h.userNameMinLength || len(userName) > h.userNameMaxLength {
		return utils.NewValidationError(constants.UserNameFormField + ": " +
			"username must be between 1 and 100 characters")
	}

	h.LogInfo("User %s (Telegram ID: %d) is changing username to %s",
		user.Name, user.TelegramID, userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.DB.Users.Update(updatedUser); err != nil {
		return err
	}

	h.LogInfo("Successfully updated username for user %d from %s to %s",
		user.TelegramID, user.Name, userName)

	// Create feed entry
	feedTitle := "Benutzername geändert"
	feedContent := fmt.Sprintf("Alter Name: %s\nNeuer Name: %s", user.Name, userName)
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for username change: %v", err)
	}

	user.Name = userName

	return nil
}
