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
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, path+"/login", h.GetLoginPage),
		utils.NewEchoRoute(http.MethodPost, path+"/login", h.PostLoginPage),

		utils.NewEchoRoute(http.MethodGet, path+"/logout", h.GetLogout),
	})
}

func (h *Handler) GetLoginPage(c echo.Context) error {
	invalid := utils.ParseQueryBool(c, "invalid")
	apiKey := c.FormValue("api-key")
	page := templates.LoginPage(apiKey, invalid)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render login page")
	}

	return nil
}

func (h *Handler) PostLoginPage(c echo.Context) error {
	// Parse form
	apiKey := c.FormValue("api-key")
	if apiKey == "" || h.processApiKeyLogin(apiKey, c) != nil {
		invalid := true
		return utils.RedirectTo(c, utils.UrlLogin(apiKey, &invalid).Page)
	}

	slog.Info("Successful login", "real_ip", c.RealIP())
	return utils.RedirectTo(c, utils.UrlProfile("").Page)
}

func (h *Handler) GetLogout(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	if cookie, err := c.Cookie(env.CookieName); err == nil {
		if err := h.registry.Cookies.Remove(cookie.Value); err != nil {
			slog.Error("Failed to remove cookie", "user_name", user.Name, "error", err)
		}
	}

	return utils.RedirectTo(c, utils.UrlLogin("", nil).Page)
}

func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) error {
	if len(apiKey) < 16 {
		return fmt.Errorf("api key too short")
	}

	user, err := h.registry.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			return errors.Wrap(err, "database authentication")
		} else {
			return errors.Wrap(err, "invalid api key")
		}
	}

	if err := h.clearExistingSession(ctx); err != nil {
		slog.Error("Failed to clear existing session", "error", err)
	}

	if err := h.createSession(ctx, apiKey); err != nil {
		return errors.Wrap(err, "create session")
	}

	if user.IsAdmin() {
		slog.Info("Administrator logged in", "user_name", user.Name, "real_ip", ctx.RealIP())
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
	cookieValue := uuid.New().String()
	isHTTPS := ctx.Request().TLS != nil || ctx.Scheme() == "https"

	cookiePath := env.ServerPathPrefix
	if cookiePath == "" {
		cookiePath = "/"
	}

	cookie := &http.Cookie{
		Name:     env.CookieName,
		Value:    cookieValue,
		Expires:  time.Now().Add(env.CookieExpirationDuration),
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteLaxMode,
	}
	ctx.SetCookie(cookie)

	sessionCookie := models.NewCookie(ctx.Request().UserAgent(), cookieValue, apiKey)
	return h.registry.Cookies.Add(sessionCookie)
}
