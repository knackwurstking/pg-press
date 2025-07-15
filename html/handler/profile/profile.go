package profile

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/html/handler"
	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

type ProfilePageData struct {
	User    *pgvis.User
	Cookies []*pgvis.Cookie
}

func NewProfilePageData() ProfilePageData {
	return ProfilePageData{
		Cookies: make([]*pgvis.Cookie, 0),
	}
}

func (p ProfilePageData) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/profile", func(c echo.Context) error {
		pageData := NewProfilePageData()

		if user, err := handler.GetUserFromContext(c); err != nil {
			return err
		} else {
			pageData.User = user
		}

		if err := handleUserNameChange(c, &pageData, db); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("change username: %s", err.Error()),
			)
		}

		if cookies, err := db.Cookies.ListApiKey(pageData.User.ApiKey); err != nil {
			log.Error("/profile -> List cookies for Api Key failed: %s", err.Error())
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

func handleUserNameChange(ctx echo.Context, pageData *ProfilePageData, db *pgvis.DB) error {
	v, err := ctx.FormParams()
	userName := v.Get("user-name")

	// Database update
	if userName != "" && userName != pageData.User.UserName {
		log.Debugf(
			"Change user name in database: %s => %s",
			pageData.User.UserName, userName,
		)

		user := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)
		if err = db.Users.Update(pageData.User.TelegramID, user); err != nil {
			return err
		} else {
			pageData.User.UserName = userName
		}
	}

	return nil
}
