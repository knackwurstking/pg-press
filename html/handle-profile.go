package html

import (
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/pgvis"
)

type ProfilePageData struct {
	PageData

	User    *pgvis.User
	Cookies []*pgvis.Cookie
}

func NewProfilePageData() ProfilePageData {
	return ProfilePageData{
		PageData: NewPageData(),
		Cookies:  make([]*pgvis.Cookie, 0),
	}
}

func (p ProfilePageData) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

func handleProfile(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	pageData := NewProfilePageData()

	if user, err := getUserFromContext(ctx); err != nil {
		return err
	} else {
		pageData.User = user
	}

	handleProfileUserName(ctx, &pageData, db)
	handleProfileCookieDeletion(ctx, &pageData, db)

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

func handleProfileUserName(ctx echo.Context, pageData *ProfilePageData, db *pgvis.DB) {
	v, err := ctx.FormParams()
	userName := v.Get("user-name")

	// Database update
	if userName != "" && userName != pageData.User.UserName {
		log.Debugf(
			"/profile -> Change user name in database: %s => %s",
			pageData.User.UserName, userName,
		)

		user := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)
		if err = db.Users.Update(pageData.User.TelegramID, user); err != nil {
			pageData.ErrorMessages = append(pageData.ErrorMessages, err.Error())
		} else {
			pageData.User.UserName = userName
		}
	}
}

func handleProfileCookieDeletion(ctx echo.Context, pageData *ProfilePageData, db *pgvis.DB) {
	deleteCookie := ctx.QueryParam("delete_cookie")
	if deleteCookie == "" {
		return
	}

	if err := db.Cookies.Remove(deleteCookie); err != nil {
		pageData.ErrorMessages = append(pageData.ErrorMessages, err.Error())
	}
}
