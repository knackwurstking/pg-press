package profile

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

type Profile struct {
	User    *pgvis.User
	Cookies []*pgvis.Cookie
}

func (p *Profile) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/profile", func(c echo.Context) error {
		pageData := &Profile{
			Cookies: []*pgvis.Cookie{},
		}

		if user, err := utils.GetUserFromContext(c); err != nil {
			return err
		} else {
			pageData.User = user
		}

		if err := handleUserNameChange(c, pageData, db); err != nil {
			return utils.HandlePgvisError(c, err)
		}

		if cookies, err := db.Cookies.ListApiKey(pageData.User.ApiKey); err != nil {
			log.Errorf("/profile -> List cookies for Api Key failed: %v", err)
		} else {
			pageData.Cookies = cookies
		}

		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/profile.html",
			"templates/nav/feed.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), pageData)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	e.GET(serverPathPrefix+"/profile/cookies", func(c echo.Context) error {
		return GETCookies(templates, c, db)
	})

	e.DELETE(serverPathPrefix+"/profile/cookies", func(c echo.Context) error {
		return DELETECookies(templates, c, db)
	})
}

func handleUserNameChange(ctx echo.Context, pageData *Profile, db *pgvis.DB) error {
	v, _ := ctx.FormParams()
	userName := utils.SanitizeInput(v.Get("user-name"))

	// Validate new username if provided
	if userName != "" && userName != pageData.User.UserName {
		// Validate username length
		if len(userName) < 1 || len(userName) > 100 {
			return pgvis.NewValidationError("user-name",
				"username must be between 1 and 100 characters", len(userName))
		}

		log.Debugf(
			"Change user name in database: %s => %s",
			pageData.User.UserName, userName,
		)

		user := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)
		if err := db.Users.Update(pageData.User.TelegramID, user); err != nil {
			return err
		}

		pageData.User.UserName = userName
	}

	return nil
}
