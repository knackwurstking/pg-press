package profile

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/profile/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetProfilePage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = handleUserNameChange(c, user)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(templates.PageProps{User: user})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Profile Page")
	}

	return nil
}

func handleUserNameChange(c echo.Context, user *shared.User) *errors.MasterError {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < shared.UserNameMinLength || len(userName) > shared.UserNameMaxLength {
		return errors.NewValidationError(
			"username must be between %d and %d characters",
			shared.UserNameMinLength, shared.UserNameMaxLength,
		).MasterError()
	}

	user.Name = userName
	merr := db.UpdateUser(user)
	if merr != nil {
		return merr
	}
	log.Info("User %#v changed their name to %#v\n", user.Name, userName)

	return nil
}
