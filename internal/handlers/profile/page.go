package profile

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/profile/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetProfilePage(c echo.Context) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr.Echo()
	}

	t := templates.Page(templates.PageProps{User: user})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Profile Page")
	}

	return nil
}

func PostProfilePage(c echo.Context) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr.Echo()
	}

	if herr = handleUserNameChange(c, user); herr != nil {
		return herr.Echo()
	}

	if herr = utils.RedirectTo(c, urlb.Profile()); herr != nil {
		return herr.Echo()
	}

	return nil
}

func handleUserNameChange(c echo.Context, user *shared.User) *errors.HTTPError {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < shared.UserNameMinLength || len(userName) > shared.UserNameMaxLength {
		return errors.NewValidationError(
			"username must be between %d and %d characters",
			shared.UserNameMinLength, shared.UserNameMaxLength,
		).HTTPError()
	}

	user.Name = userName
	herr := db.UpdateUser(user)
	if herr != nil {
		return herr
	}
	log.Info("User %#v changed their name to %#v\n", user.Name, userName)

	return nil
}
