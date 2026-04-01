package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/auth/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func GetLoginPage(c echo.Context) *echo.HTTPError {
	slog.Debug("Login page requested from IP", "real_ip", c.RealIP())

	t := templates.Page(
		templates.PageProps{
			FormData: map[string]string{
				"api-key":         c.FormValue("api-key"),
				"invalid-api-key": fmt.Sprintf("%t", utils.GetQueryBool(c, "invalid")),
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
	slog.Debug("Login attempt from IP", "real_ip", c.RealIP())

	apiKey := c.FormValue("api-key")

	err := processApiKeyLogin(apiKey, c)
	if err != nil {
		slog.Warn("Login failed", "error", err)
	}
	if apiKey == "" || err != nil {
		invalid := true
		merr := utils.RedirectTo(c, urlb.Login(apiKey, &invalid))
		if merr != nil {
			return merr.Echo()
		}
	}

	merr := utils.RedirectTo(c, urlb.Profile())
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func processApiKeyLogin(apiKey string, ctx echo.Context) *errors.HTTPError {
	if len(apiKey) < shared.MinAPIKeyLength {
		return errors.NewValidationError("invalid API key: too short").HTTPError()
	}

	user, merr := db.GetUserByApiKey(apiKey)
	if merr != nil {
		return merr
	}

	merr = createSession(ctx, user.ID)
	if merr != nil {
		return merr
	}

	if user.IsAdmin() {
		slog.Info("Administrator logged in",
			"user_name", user.Name,
			"real_ip", ctx.RealIP())
	}

	return nil
}

func createSession(ctx echo.Context, userID shared.TelegramID) *errors.HTTPError {
	cookie := &shared.Cookie{
		UserAgent: ctx.Request().UserAgent(),
		Value:     uuid.New().String(),
		UserID:    userID,
		LastLogin: shared.NewUnixMilli(time.Now()),
	}

	if cookieContext, _ := ctx.Cookie(CookieName); cookieContext != nil && cookieContext.Value != "" {
		merr := db.DeleteCookie(cookie.Value)
		if merr != nil {
			return merr
		}
	}

	merr := db.AddCookie(cookie)
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
