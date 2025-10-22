package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	*Base
}

func NewAuth(db *services.Registry) *Auth {
	return &Auth{
		Base: NewBase(db, logger.NewComponentLogger("Auth")),
	}
}

func (h *Auth) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/login", h.GetLoginPage),
		utils.NewEchoRoute(http.MethodGet, "/logout", h.GetLogout),
	})
}

func (h *Auth) GetLoginPage(c echo.Context) error {
	apiKey := c.FormValue("api-key")
	if apiKey != "" {
		if h.processApiKeyLogin(apiKey, c) {
			h.Log.Info("Successful login from %s", c.RealIP())
			return RedirectTo(c, "./profile")
		}
		h.Log.Info("Failed login attempt from %s", c.RealIP())
	}

	page := components.PageLogin(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render login page")
	}

	return nil
}

func (h *Auth) GetLogout(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	if cookie, err := c.Cookie(env.CookieName); err == nil {
		if err := h.Registry.Cookies.Remove(cookie.Value); err != nil {
			h.Log.Error("Failed to remove cookie for user %s: %v", user.Name, err)
		}
	}

	return RedirectTo(c, "./login")
}

func (h *Auth) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	if len(apiKey) < 16 {
		h.Log.Debug("API key too short from %s", ctx.RealIP())
		return false
	}

	user, err := h.Registry.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			h.Log.Error("Database error during authentication: %v", err)
		} else {
			h.Log.Debug("Invalid API key from %s", ctx.RealIP())
		}
		return false
	}

	if err := h.clearExistingSession(ctx, user.Name); err != nil {
		h.Log.Error("Failed to clear existing session: %v", err)
	}

	if err := h.createSession(ctx, apiKey, user); err != nil {
		h.Log.Error("Failed to create session for user %s: %v", user.Name, err)
		return false
	}

	if user.IsAdmin() {
		h.Log.Info("Administrator %s logged in from %s", user.Name, ctx.RealIP())
	}

	return true
}

func (h *Auth) clearExistingSession(ctx echo.Context, username string) error {
	cookie, err := ctx.Cookie(env.CookieName)
	if err != nil {
		return nil
	}

	if err := h.Registry.Cookies.Remove(cookie.Value); err != nil {
		return err
	}

	return nil
}

func (h *Auth) createSession(ctx echo.Context, apiKey string, user *models.User) error {
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
	return h.Registry.Cookies.Add(sessionCookie)
}
