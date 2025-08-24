package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Profile struct {
	DB *database.DB
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	prefix := "/profile"

	e.GET(serverPathPrefix+prefix, h.handleMainPage)
	e.GET(serverPathPrefix+prefix+"/", h.handleMainPage)

	htmxProfile := htmxhandler.Profile{DB: h.DB}
	htmxProfile.RegisterRoutes(e)
}

func (h *Profile) handleMainPage(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HandlerProfile().Debug("Rendering profile page for user %s", user.UserName)

	if err = h.handleUserNameChange(c, user); err != nil {
		logger.HandlerProfile().Error("Failed to update username for user %s: %v", user.UserName, err)
		return echo.NewHTTPError(
			database.GetHTTPStatusCode(err),
			"error updating username: "+err.Error(),
		)
	}

	page := pages.ProfilePage(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerProfile().Error("Failed to render profile page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render profile page: "+err.Error())
	}
	return nil
}

func (h *Profile) handleUserNameChange(c echo.Context, user *database.User) error {
	formParams, _ := c.FormParams()
	userName := utils.SanitizeInput(formParams.Get(constants.UserNameFormField))

	if userName == "" || userName == user.UserName {
		return nil
	}

	if len(userName) < constants.UserNameMinLength || len(userName) > constants.UserNameMaxLength {
		logger.HandlerProfile().Warn("Invalid username length for user %s: %d characters (attempted: %s)",
			user.UserName, len(userName), userName)
		return database.NewValidationError(constants.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	logger.HandlerProfile().Info("User %s (Telegram ID: %d) is changing username to %s",
		user.UserName, user.TelegramID, userName)

	updatedUser := database.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.DB.Users.Update(user.TelegramID, updatedUser); err != nil {
		logger.HandlerProfile().Error("Failed to update username in database: %v", err)
		return err
	}

	logger.HandlerProfile().Info("Successfully updated username for user %d from %s to %s",
		user.TelegramID, user.UserName, userName)
	user.UserName = userName
	return nil
}
