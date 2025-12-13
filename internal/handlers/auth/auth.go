package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/auth/templates"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgpress-api-key"
)

type Handler struct {
	db *common.DB
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path+"/login", h.GetLoginPage),
		ui.NewEchoRoute(http.MethodPost, path+"/login", h.PostLoginPage),

		ui.NewEchoRoute(http.MethodGet, path+"/logout", h.GetLogout),
	})
}

func (h *Handler) GetLoginPage(c echo.Context) error {
	var (
		invalid = shared.ParseQueryBool(c, "invalid")
		apiKey  = c.FormValue("api-key")
		page    = templates.LoginPage(apiKey, invalid)
	)

	err := page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Login Page")
	}

	return nil
}

func (h *Handler) PostLoginPage(c echo.Context) error {
	slog.Info("User authentication request received")

	// Parse form
	apiKey := c.FormValue("api-key")
	err := h.processApiKeyLogin(apiKey, c)
	if err != nil {
		slog.Warn("Processing api key failed", "api_key", shared.MaskString(apiKey), "error", err)
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

func (h *Handler) GetLogout(c echo.Context) error {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookie, err := c.Cookie(CookieName)
	if err == nil {
		merr := h.db.User.Cookie.Delete(cookie.Value)
		if merr != nil {
			slog.Error("Failed to remove cookie", "user_name", user.Name, "error", merr)
		}
	}

	merr = urlb.RedirectTo(c, urlb.UrlLogin("", nil).Page)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) error {
	if len(apiKey) < shared.MinAPIKeyLength {
		return fmt.Errorf("api key too short")
	}

	user, merr := helper.GetUserForApiKey(h.db, apiKey)
	if merr != nil {
		if merr.Code != http.StatusNotFound {
			return errors.Wrap(merr, "database authentication")
		} else {
			return errors.Wrap(merr, "invalid api key")
		}
	}

	err := h.clearExistingSession(ctx)
	if err != nil {
		slog.Warn("Failed to clear existing session", "error", err)
	}

	err = h.createSession(ctx, apiKey)
	if err != nil {
		return errors.Wrap(err, "create session")
	}

	if user.IsAdmin() {
		slog.Info("Administrator login successful", "user", user.Name, "ip", ctx.RealIP())
	}

	return nil
}

func (h *Handler) clearExistingSession(ctx echo.Context) error {
	cookie, err := ctx.Cookie(CookieName)
	if err != nil {
		return nil
	}

	if err := h.registry.Cookies.Remove(cookie.Value); err != nil {
		return err
	}

	return nil
}

func (h *Handler) createSession(ctx echo.Context, apiKey string) error {
	var (
		cookieValue = uuid.New().String()
		cookiePath  = env.ServerPathPrefix
	)

	if cookiePath == "" {
		cookiePath = "/"
	}

	err := h.registry.Cookies.Add(ctx.Request().UserAgent(), cookieValue, apiKey)
	if err != nil {
		return err
	}

	ctx.SetCookie(&http.Cookie{
		Name:     CookieName,
		Value:    cookieValue,
		Expires:  time.UnixMilli(time.Now().UnixMilli() + shared.CookieExpirationDuration), // TODO: Use the cookie's ExpiredAt method
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   ctx.Request().TLS != nil || ctx.Scheme() == "https",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
