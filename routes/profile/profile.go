// Package profile provides HTTP route handlers for user profile management.
package profile

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// Profile contains the data structure passed to the profile page template.
type Profile struct {
	User    *pgvis.User     `json:"user"`
	Cookies []*pgvis.Cookie `json:"cookies"`
}

// CookiesSorted returns the user's cookies sorted by last login time.
func (p *Profile) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

// Serve configures and registers all profile related HTTP routes.
func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/profile", handleMainPage(templates, db))

	cookiesPath := serverPathPrefix + "/profile/cookies"
	e.GET(cookiesPath, handleGetCookies(templates, db))
	e.DELETE(cookiesPath, handleDeleteCookies(templates, db))
}

func handleMainPage(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		pageData := &Profile{
			Cookies: make([]*pgvis.Cookie, 0),
		}

		user, herr := utils.GetUserFromContext(c)
		if herr != nil {
			return herr
		}
		pageData.User = user

		if err := handleUserNameChange(c, pageData, db); err != nil {
			return utils.HandlePgvisError(c, err)
		}

		if cookies, err := db.Cookies.ListApiKey(pageData.User.ApiKey); err == nil {
			pageData.Cookies = cookies
		}

		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.ProfileTemplatePath,
			shared.NavFeedTemplatePath,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				pgvis.WrapError(err, "failed to parse template"))
		}

		if err = t.Execute(c.Response(), pageData); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				pgvis.WrapError(err, "failed to execute template"))
		}

		return nil
	}
}

func handleGetCookies(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETCookies(templates, c, db)
	}
}

func handleDeleteCookies(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return DELETECookies(templates, c, db)
	}
}

func handleUserNameChange(ctx echo.Context, pageData *Profile, db *pgvis.DB) error {
	formParams, _ := ctx.FormParams()
	userName := utils.SanitizeInput(formParams.Get(shared.UserNameFormField))

	if userName == "" || userName == pageData.User.UserName {
		return nil
	}

	if len(userName) < shared.UserNameMinLength || len(userName) > shared.UserNameMaxLength {
		return pgvis.NewValidationError(shared.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	updatedUser := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)

	if err := db.Users.Update(pageData.User.TelegramID, updatedUser); err != nil {
		return err
	}

	pageData.User.UserName = userName
	return nil
}
