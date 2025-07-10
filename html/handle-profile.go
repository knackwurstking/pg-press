package html

import (
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

type ProfilePageData struct {
	PageData

	User    *pgvis.User
	Cookies []*pgvis.Cookie
}

func (p ProfilePageData) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

func handleProfile(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	pageData := ProfilePageData{
		Cookies: make([]*pgvis.Cookie, 0),
	}

	if user, err := getUserFromContext(ctx); err != nil {
		return err
	} else {
		pageData.User = user
	}

	// Get "user-name" from form data (optional), and update database user
	v, err := ctx.FormParams()
	userName := v.Get("user-name")

	// Database update
	if userName != "" && userName != pageData.User.UserName {
		log.Debugf(
			"/profile -> Change user name in database: %s => %s",
			pageData.User.UserName, userName,
		)

		pageData.User.UserName = userName
		if err = db.Users.Update(pageData.User.TelegramID, pageData.User); err != nil {
			pageData.ErrorMessages = []string{err.Error()}
		}
	}

	if cookies, err := db.Cookies.ListApiKey(pageData.User.ApiKey); err != nil {
		log.Error("/profile -> List cookies for Api Key failed: %s", err.Error())
	} else {
		pageData.Cookies = cookies
	}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/profile.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(ctx.Response(), pageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
