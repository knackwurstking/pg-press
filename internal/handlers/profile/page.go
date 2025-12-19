package profile

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/errors"
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

	t := Page(PageProps{User: user})
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
		return errors.NewMasterError(
			fmt.Errorf("username must be between 1 and 100 characters"),
			http.StatusBadRequest,
		)
	}

	user.Name = userName
	merr := DB.User.Users.Update(user)
	if merr != nil {
		return merr
	}
	Log.Info("User %#v changed their name to %#v\n", user.Name, userName)

	return nil
}
