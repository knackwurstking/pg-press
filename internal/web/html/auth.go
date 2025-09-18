package html

import (
	"net/http"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/loginpage"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	*handlers.BaseHandler
}

func NewAuth(db *database.DB) *Auth {
	return &Auth{
		BaseHandler: handlers.NewBaseHandler(db, logger.HandlerAuth()),
	}
}

func (h *Auth) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/login", h.HandleLogin),
			helpers.NewEchoRoute(http.MethodGet, "/logout", h.HandleLogout),
		},
	)
}

// handleLogin handles the login page and form submission.
func (h *Auth) HandleLogin(c echo.Context) error {
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

	page := loginpage.Page(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render login page: "+err.Error())
	}

	return nil
}

// handleLogout handles user logout.
func (h *Auth) HandleLogout(c echo.Context) error {
	remoteIP := c.RealIP()

	// Try to get user info before logout for better logging
	user, err := h.GetUserFromContext(c)
	if err == nil {
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
func (h *Auth) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
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
	cookie := &http.Cookie{
		Name:     constants.CookieName,
		Value:    cookieValue,
		Expires:  time.Now().Add(constants.CookieExpirationDuration),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	ctx.SetCookie(cookie)

	sessionCookie := models.NewCookie(userAgent, cookieValue, apiKey)
	if err := h.DB.Cookies.Add(sessionCookie); err != nil {
		return false
	}

	// Log security-relevant information
	if user.IsAdmin() {
		h.LogInfo("Administrator %s logged in from %s", user.Name, remoteIP)
	}

	return true
}
