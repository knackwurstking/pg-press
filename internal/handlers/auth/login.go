package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func GetLoginPage(c echo.Context) *echo.HTTPError {
	Log.Debug("Login page requested from IP: %#v", c.RealIP())

	t := Page(
		PageProps{
			FormData: map[string]string{
				"api-key":         c.FormValue("api-key"),
				"invalid-api-key": fmt.Sprintf("%t", shared.ParseQueryBool(c, "invalid")),
			},
		},
	)

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Login Page")
	}
	return nil
}

func PostLoginPage(c echo.Context) *echo.HTTPError {
	Log.Debug("Login attempt from IP: %#v", c.RealIP())

	apiKey := c.FormValue("api-key")

	err := processApiKeyLogin(apiKey, c)
	if err != nil {
		fmt.Println("Processing api key failed:", err)
	}
	if apiKey == "" || err != nil {
		invalid := true
		merr := urlb.RedirectTo(c, urlb.UrlLogin(apiKey, &invalid).Page)
		if merr != nil {
			return merr.Echo()
		}
	}

	merr := urlb.RedirectTo(c, urlb.UrlProfile("").Page)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func processApiKeyLogin(apiKey string, ctx echo.Context) *errors.MasterError {
	if len(apiKey) < shared.MinAPIKeyLength {
		return errors.NewMasterError(fmt.Errorf("api key too short"), http.StatusBadRequest)
	}

	user, merr := helper.GetUserForApiKey(DB, apiKey)
	if merr != nil {
		return merr
	}

	merr = clearExistingSession(ctx)
	if merr != nil {
		Log.Warn("Failed to clear existing session: %v", merr)
	}

	merr = createSession(ctx, user.ID)
	if merr != nil {
		return merr
	}

	if user.IsAdmin() {
		Log.Info("Administrator login successful: %#v, from ID: %#v", user.Name, ctx.RealIP())
	}

	return nil
}

func clearExistingSession(ctx echo.Context) *errors.MasterError {
	cookie, err := ctx.Cookie(CookieName)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	merr := DB.User.Cookie.Delete(cookie.Value)
	if merr != nil {
		return merr
	}

	return nil
}

func createSession(ctx echo.Context, userID shared.TelegramID) *errors.MasterError {
	cookie := &shared.Cookie{
		UserAgent: ctx.Request().UserAgent(),
		Value:     uuid.New().String(),
		UserID:    userID,
		LastLogin: shared.NewUnixMilli(time.Now()),
	}
	merr := DB.User.Cookie.Create(cookie)
	if merr != nil {
		return merr
	}

	ctx.SetCookie(&http.Cookie{
		Name:     CookieName,
		Value:    cookie.Value,
		Expires:  cookie.ExpiredAtTime(),
		Path:     env.ServerPathPrefix,
		HttpOnly: true,
		Secure:   ctx.Request().TLS != nil || ctx.Scheme() == "https",
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}
