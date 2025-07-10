package html

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/pgvis"
)

type LoginPageData struct {
	PageData

	ApiKey        string
	InvalidApiKey bool
}

func NewLoginPageData() LoginPageData {
	return LoginPageData{
		PageData: NewPageData(),
	}
}

func handleLogin(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	v, err := ctx.FormParams()
	apiKey := v.Get("api-key")

	if ok, err := handleLoginApiKey(apiKey, db, ctx); ok {
		if err = ctx.Redirect(http.StatusSeeOther, "./profile"); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		} else {
			return nil
		}
	} else {
		if err != nil {
			log.Errorf("/login -> Invalid Api Key: %s", err.Error())
		}
	}

	pageData := NewLoginPageData()
	pageData.ApiKey = apiKey
	pageData.InvalidApiKey = apiKey != ""

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/login.html",
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

func handleLoginApiKey(apiKey string, db *pgvis.DB, ctx echo.Context) (ok bool, err error) {
	if apiKey == "" {
		return false, nil
	}

	u, err := db.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		return false, fmt.Errorf("database error: %s", err.Error())
	}

	// Set cookie and redirect to "/"
	if u.ApiKey == apiKey {
		if cookie, err := ctx.Cookie(CookieName); err == nil {
			log.Debug("HandleLoginApiKey -> Removing the old cookie...")
			if err = db.Cookies.Remove(cookie.Value); err != nil {
				log.Warnf("HandleLoginApiKey -> Removing the old cookie failed: %s", err.Error())
			}
		}

		log.Debugf(
			"HandleLoginApiKey -> Set cookie and redirect to /profile (id: %#v; user: %#v)",
			u.TelegramID, u.UserName,
		)

		cookie := new(http.Cookie)

		cookie.Name = CookieName
		cookie.Value = uuid.New().String()
		ctx.SetCookie(cookie)

		db.Cookies.Add(&pgvis.Cookie{
			UserAgent: ctx.Request().UserAgent(),
			Value:     cookie.Value,
			ApiKey:    apiKey,
			LastLogin: time.Now().UnixMilli(),
		})

		return true, nil
	}

	return false, nil
}
