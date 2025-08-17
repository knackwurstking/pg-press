package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Profile struct {
	DB *database.DB
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	prefix := "/profile"

	e.GET(serverPathPrefix+prefix, h.handleMainPage)

	htmxProfile := htmxhandler.Profile{DB: h.DB}
	htmxProfile.RegisterRoutes(e)
}

func (h *Profile) handleMainPage(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	if err = h.handleUserNameChange(c, user); err != nil {
		return echo.NewHTTPError(
			database.GetHTTPStatusCode(err),
			"error updating username: "+err.Error(),
		)
	}

	page := pages.ProfilePage(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
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
		return database.NewValidationError(constants.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	updatedUser := database.NewUser(user.TelegramID, userName, user.ApiKey)
	if err := h.DB.Users.Update(user.TelegramID, updatedUser); err != nil {
		return err
	}

	user.UserName = userName
	return nil
}
