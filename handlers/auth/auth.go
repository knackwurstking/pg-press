package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/auth/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
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
		invalid = utils.ParseQueryBool(c, "invalid")
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
		slog.Warn("Processing api key failed", "api_key", utils.MaskString(apiKey), "error", err)
	}
	if apiKey == "" || err != nil {
		invalid := true
		merr := utils.RedirectTo(c, utils.UrlLogin(apiKey, &invalid).Page)
		if merr != nil {
			return merr.Echo()
		}
	}

	merr := utils.RedirectTo(c, utils.UrlProfile("").Page)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func (h *Handler) GetLogout(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookie, err := c.Cookie(env.CookieName)
	if err == nil {
		merr := h.registry.Cookies.Remove(cookie.Value)
		if merr != nil {
			slog.Error("Failed to remove cookie", "user_name", user.Name, "error", merr)
		}
	}

	merr = utils.RedirectTo(c, utils.UrlLogin("", nil).Page)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) error {
	if len(apiKey) < env.MinAPIKeyLength {
		return fmt.Errorf("api key too short")
	}

	user, merr := h.registry.Users.GetUserFromApiKey(apiKey)
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
	cookie, err := ctx.Cookie(env.CookieName)
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

	err := h.registry.Cookies.Add(
		models.NewCookie(ctx.Request().UserAgent(), cookieValue, apiKey),
	)
	if err != nil {
		return err
	}

	ctx.SetCookie(&http.Cookie{
		Name:     env.CookieName,
		Value:    cookieValue,
		Expires:  time.Now().Add(env.CookieExpirationDuration),
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   ctx.Request().TLS != nil || ctx.Scheme() == "https",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
