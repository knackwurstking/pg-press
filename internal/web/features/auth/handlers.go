package auth

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/web/features/auth/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(db, logger.NewComponentLogger("Auth")),
	}
}

func (h *Handler) LoginPage(c echo.Context) error {
	remoteIP := c.RealIP()

	apiKey := h.GetSanitizedFormValue(c, constants.APIKeyFormField)
	if apiKey != "" {
		if h.processApiKeyLogin(apiKey, c) {
			h.LogInfo("Successful login for user from %s", remoteIP)
			if err := h.RedirectTo(c, "./profile"); err != nil {
				return h.RenderInternalError(c,
					"failed to redirect to profile page: "+err.Error())
			}
			return nil
		} else {
			h.LogInfo("Failed login attempt from %s", remoteIP)
		}
	}

	page := templates.LoginPage(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render login page: "+err.Error())
	}

	return nil
}

// handleLogout handles user logout.
func (h *Handler) Logout(c echo.Context) error {
	remoteIP := c.RealIP()

	// Try to get user info before logout for better logging
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	if cookie, err := c.Cookie(constants.CookieName); err == nil {
		if err := h.DB.Cookies.Remove(cookie.Value); err != nil {
			h.LogError("Failed to remove cookie from database for user %s from %s: %v",
				user.Name, remoteIP, err)
		}
	}

	if err := h.RedirectTo(c, "./login"); err != nil {
		return h.RenderInternalError(c,
			"failed to redirect to login page: "+err.Error())
	}

	return nil
}

// processApiKeyLogin processes API key authentication and creates a session.
func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	remoteIP := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	// Validate API key format
	if len(apiKey) < 16 {
		h.LogDebug("API key too short from %s", remoteIP)
		return false
	}

	user, err := h.DB.Users.GetUserFromApiKey(apiKey)

	if err != nil {
		if !utils.IsNotFoundError(err) {
			h.LogError("Database error during authentication from %s: %v", remoteIP, err)
		} else {
			h.LogDebug("Invalid API key from %s", remoteIP)
		}

		return false
	}

	// Check for existing session
	if existingCookie, err := ctx.Cookie(constants.CookieName); err == nil {
		if err := h.DB.Cookies.Remove(existingCookie.Value); err != nil {
			h.LogError("Failed to remove existing cookie for user %s from %s: %v",
				user.Name, remoteIP, err)
		}
	}

	// Create new session cookie
	cookieValue := uuid.New().String()

	// Only set Secure flag for HTTPS requests (PWA compatibility)
	isHTTPS := ctx.Request().TLS != nil || ctx.Scheme() == "https"

	// Use ServerPathPrefix or fallback to root for cookie path
	cookiePath := env.ServerPathPrefix
	if cookiePath == "" {
		cookiePath = "/"
	}

	// Log cookie creation details for debugging PWA issues
	h.LogDebug("Creating cookie for user %s from %s: HTTPS=%v, UserAgent='%s', Scheme='%s', Path='%s'",
		user.Name, remoteIP, isHTTPS, userAgent, ctx.Scheme(), cookiePath)

	cookie := &http.Cookie{
		Name:     constants.CookieName,
		Value:    cookieValue,
		Expires:  time.Now().Add(constants.CookieExpirationDuration),
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteLaxMode,
	}
	ctx.SetCookie(cookie)

	h.LogDebug("Cookie set successfully: Name=%s, Secure=%v, SameSite=%v, Path=%s, Expires=%v",
		cookie.Name, cookie.Secure, cookie.SameSite, cookie.Path, cookie.Expires)

	sessionCookie := models.NewCookie(userAgent, cookieValue, apiKey)
	if err := h.DB.Cookies.Add(sessionCookie); err != nil {
		h.LogError("Failed to store session cookie for user %s from %s: %v", user.Name, remoteIP, err)
		return false
	}

	h.LogDebug("Session cookie stored in database for user %s", user.Name)

	// Log security-relevant information
	if user.IsAdmin() {
		h.LogInfo("Administrator %s logged in from %s", user.Name, remoteIP)
	}

	return true
}
