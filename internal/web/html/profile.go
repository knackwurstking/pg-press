package html

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
	"github.com/knackwurstking/pgpress/internal/web/constants"

	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	profilepage "github.com/knackwurstking/pgpress/internal/web/templates/pages/profile"

	"github.com/labstack/echo/v4"
)

const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
)

type Profile struct {
	DB *database.DB
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/profile", h.handleProfile),
		},
	)
}

func (h *Profile) handleProfile(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HandlerProfile().Debug("Rendering profile page for user %s", user.Name)

	if err = h.handleUserNameChange(c, user); err != nil {
		logger.HandlerProfile().Error("Failed to update username for user %s: %v", user.Name, err)
		return echo.NewHTTPError(
			dberror.GetHTTPStatusCode(err),
			"error updating username: "+err.Error(),
		)
	}

	page := profilepage.Page(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerProfile().Error("Failed to render profile page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render profile page: "+err.Error())
	}
	return nil
}

func (h *Profile) handleUserNameChange(c echo.Context, user *models.User) error {
	formParams, _ := c.FormParams()
	userName := webhelpers.SanitizeInput(formParams.Get(constants.UserNameFormField))

	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < UserNameMinLength || len(userName) > UserNameMaxLength {
		logger.HandlerProfile().Warn("Invalid username length for user %s: %d characters (attempted: %s)",
			user.Name, len(userName), userName)
		return dberror.NewValidationError(constants.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	logger.HandlerProfile().Info("User %s (Telegram ID: %d) is changing username to %s",
		user.Name, user.TelegramID, userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.DB.Users.Update(updatedUser, user); err != nil {
		logger.HandlerProfile().Error("Failed to update username in database: %v", err)
		return err
	}

	logger.HandlerProfile().Info("Successfully updated username for user %d from %s to %s",
		user.TelegramID, user.Name, userName)
	user.Name = userName
	return nil
}
