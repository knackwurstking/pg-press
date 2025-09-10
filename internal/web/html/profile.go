package html

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/profilepage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

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
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/profile", h.handleProfile),
		},
	)
}

func (h *Profile) handleProfile(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HandlerProfile().Debug("Rendering profile page for user %s", user.Name)

	if err = h.handleUserNameChange(c, user); err != nil {
		logger.HandlerProfile().Error("Failed to update username for user %s: %v", user.Name, err)
		return echo.NewHTTPError(
			utils.GetHTTPStatusCode(err),
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
	userName := helpers.SanitizeInput(formParams.Get(constants.UserNameFormField))

	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < UserNameMinLength || len(userName) > UserNameMaxLength {
		return utils.NewValidationError(
			constants.UserNameFormField + ": " +
				"username must be between 1 and 100 characters",
		)
	}

	logger.HandlerProfile().Info("User %s (Telegram ID: %d) is changing username to %s",
		user.Name, user.TelegramID, userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.DB.Users.Update(updatedUser, user); err != nil {
		return err
	}

	logger.HandlerProfile().Info("Successfully updated username for user %d from %s to %s",
		user.TelegramID, user.Name, userName)
	user.Name = userName
	return nil
}
