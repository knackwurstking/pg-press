package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/auth/components"
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

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/login", h.GetLoginPage),
		utils.NewEchoRoute(http.MethodGet, "/logout", h.GetLogout),
	})
}

func (h *Handler) GetLoginPage(c echo.Context) error {
	apiKey := c.FormValue("api-key")
	if apiKey != "" {
		if h.processApiKeyLogin(apiKey, c) {
			slog.Info("Successful login", "real_ip", c.RealIP())
			return utils.RedirectTo(c, "./profile")
		}
		slog.Info("Failed login attempt", "real_ip", c.RealIP())
	}

	page := components.PageLogin(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "failed to render login page")
	}

	return nil
}

func (h *Handler) GetLogout(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "failed to get user from context")
	}

	if cookie, err := c.Cookie(env.CookieName); err == nil {
		if err := h.registry.Cookies.Remove(cookie.Value); err != nil {
			slog.Error("Failed to remove cookie", "user_name", user.Name, "error", err)
		}
	}

	return utils.RedirectTo(c, "./login")
}

func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	if len(apiKey) < 16 {
		slog.Debug("API key too short", "real_ip", ctx.RealIP())
		return false
	}

	user, err := h.registry.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			slog.Error("Database error during authentication", "error", err)
		} else {
			slog.Debug("Invalid API key", "real_ip", ctx.RealIP())
		}
		return false
	}

	if err := h.clearExistingSession(ctx, user.Name); err != nil {
		slog.Error("Failed to clear existing session", "error", err)
	}

	if err := h.createSession(ctx, apiKey, user); err != nil {
		slog.Error("Failed to create session", "user_name", user.Name, "error", err)
		return false
	}

	if user.IsAdmin() {
		slog.Info("Administrator logged in", "user_name", user.Name, "real_ip", ctx.RealIP())
	}

	return true
}

func (h *Handler) clearExistingSession(ctx echo.Context, username string) error {
	cookie, err := ctx.Cookie(env.CookieName)
	if err != nil {
		return nil
	}

	if err := h.registry.Cookies.Remove(cookie.Value); err != nil {
		return err
	}

	return nil
}

func (h *Handler) createSession(ctx echo.Context, apiKey string, user *models.User) error {
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
