package html

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	cookiemodels "github.com/knackwurstking/pgpress/internal/database/models/cookie"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	loginpage "github.com/knackwurstking/pgpress/internal/web/templates/pages/login"
)

type Auth struct {
	DB *database.DB
}

func (h *Auth) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/login", h.handleLogin),
			webhelpers.NewEchoRoute(http.MethodGet, "/logout", h.handleLogout),
		},
	)
}

// handleLogin handles the login page and form submission.
func (h *Auth) handleLogin(c echo.Context) error {
	remoteIP := c.RealIP()
	userAgent := c.Request().UserAgent()
	referer := c.Request().Header.Get("Referer")

	start := time.Now()
	formParams, _ := c.FormParams()
	apiKey := strings.TrimSpace(formParams.Get(constants.APIKeyFormField))

	logger.HandlerAuth().Info("Login page request from %s (user-agent: %s, referer: %s)",
		remoteIP, userAgent, referer)

	if apiKey != "" {
		logger.HandlerAuth().Info("Processing login attempt from %s with API key (length: %d)",
			remoteIP, len(apiKey))

		loginStart := time.Now()
		if h.processApiKeyLogin(apiKey, c) {
			loginElapsed := time.Since(loginStart)
			totalElapsed := time.Since(start)
			logger.HandlerAuth().Info("Successful login from %s in %v (auth: %v, total: %v)",
				remoteIP, totalElapsed, loginElapsed, totalElapsed)

			if err := c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
				logger.HandlerAuth().Error("Failed to redirect after successful login from %s: %v", remoteIP, err)
				return echo.NewHTTPError(http.StatusInternalServerError,
					"failed to redirect to profile page")
			}
			return nil
		} else {
			loginElapsed := time.Since(loginStart)
			logger.HandlerAuth().Warn("Failed login attempt from %s in %v (invalid API key length: %d)",
				remoteIP, loginElapsed, len(apiKey))
		}
	}

	renderStart := time.Now()
	page := loginpage.Page(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerAuth().Error("Failed to render login page for %s: %v", remoteIP, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render login page: "+err.Error())
	}

	renderElapsed := time.Since(renderStart)
	totalElapsed := time.Since(start)
	logger.HandlerAuth().Debug("Login page rendered for %s (render: %v, total: %v)",
		remoteIP, renderElapsed, totalElapsed)

	return nil
}

// handleLogout handles user logout.
func (h *Auth) handleLogout(c echo.Context) error {
	remoteIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	start := time.Now()
	logger.HandlerAuth().Info("Logout request from %s (user-agent: %s)", remoteIP, userAgent)

	// Try to get user info before logout for better logging
	var userName string = "unknown"
	if user, err := webhelpers.GetUserFromContext(c); err == nil {
		userName = user.Name
		logger.HandlerAuth().Info("User %s (ID: %d) logging out from %s", user.Name, user.TelegramID, remoteIP)
	}

	cookieRemoved := false
	if cookie, err := c.Cookie(constants.CookieName); err == nil {
		logger.HandlerAuth().Debug("Removing authentication cookie for user %s from %s", userName, remoteIP)
		if err := h.DB.Cookies.Remove(cookie.Value); err != nil {
			logger.HandlerAuth().Error("Failed to remove cookie from database for user %s from %s: %v",
				userName, remoteIP, err)
		} else {
			cookieRemoved = true
			logger.HandlerAuth().Debug("Successfully removed cookie for user %s", userName)
		}
	} else {
		logger.HandlerAuth().Debug("No authentication cookie found for logout from %s", remoteIP)
	}

	if err := c.Redirect(http.StatusSeeOther, "./login"); err != nil {
		logger.HandlerAuth().Error("Failed to redirect to login page after logout from %s: %v", remoteIP, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to redirect to login page")
	}

	elapsed := time.Since(start)
	logger.HandlerAuth().Info("User %s logout completed from %s in %v (cookie removed: %v)",
		userName, remoteIP, elapsed, cookieRemoved)

	return nil
}

// processApiKeyLogin processes API key authentication and creates a session.
func (h *Auth) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	remoteIP := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	start := time.Now()

	// Validate API key format
	if len(apiKey) < 16 {
		logger.HandlerAuth().Warn("API key too short from %s: %d characters", remoteIP, len(apiKey))
		return false
	}

	dbStart := time.Now()
	user, err := h.DB.Users.GetUserFromApiKey(apiKey)
	dbElapsed := time.Since(dbStart)

	if err != nil {
		if errors.Is(err, dberror.ErrNotFound) {
			logger.HandlerAuth().Warn("Authentication failed from %s: invalid API key (db lookup: %v)",
				remoteIP, dbElapsed)
		} else {
			logger.HandlerAuth().Error("Database error during authentication from %s (db lookup: %v): %v",
				remoteIP, dbElapsed, err)
		}
		return false
	}

	logger.HandlerAuth().Info("User %s (ID: %d) authenticated from %s (db lookup: %v)",
		user.Name, user.TelegramID, remoteIP, dbElapsed)

	// Check for existing session
	if existingCookie, err := ctx.Cookie(constants.CookieName); err == nil {
		logger.HandlerAuth().Info("Removing existing authentication cookie for user %s from %s",
			user.Name, remoteIP)
		if err := h.DB.Cookies.Remove(existingCookie.Value); err != nil {
			logger.HandlerAuth().Error("Failed to remove existing cookie for user %s from %s: %v",
				user.Name, remoteIP, err)
		}
	}

	logger.HandlerAuth().Info("Creating new session for user %s (ID: %d) from %s (user-agent: %s)",
		user.Name, user.TelegramID, remoteIP, userAgent)

	// Create new session cookie
	sessionStart := time.Now()
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

	sessionCookie := cookiemodels.New(userAgent, cookieValue, apiKey)
	if err := h.DB.Cookies.Add(sessionCookie); err != nil {
		logger.HandlerAuth().Error("Failed to create session for user %s from %s: %v",
			user.Name, remoteIP, err)
		return false
	}

	sessionElapsed := time.Since(sessionStart)
	totalElapsed := time.Since(start)
	logger.HandlerAuth().Info("Successfully created session for user %s from %s in %v (db: %v, session: %v, total: %v)",
		user.Name, remoteIP, totalElapsed, dbElapsed, sessionElapsed, totalElapsed)

	// Log security-relevant information
	if user.IsAdmin() {
		logger.HandlerAuth().Info("Administrator %s logged in from %s", user.Name, remoteIP)
	}

	return true
}
