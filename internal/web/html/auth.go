package html

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/loginpage"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	DB *database.DB
}

func (h *Auth) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/login", h.handleLogin),
			helpers.NewEchoRoute(http.MethodGet, "/logout", h.handleLogout),
		},
	)
}

// handleLogin handles the login page and form submission.
func (h *Auth) handleLogin(c echo.Context) error {
	remoteIP := c.RealIP()

	formParams, _ := c.FormParams()
	apiKey := strings.TrimSpace(formParams.Get(constants.APIKeyFormField))

	if apiKey != "" {

		if h.processApiKeyLogin(apiKey, c) {
			logger.HandlerAuth().Info("Successful login for user from %s", remoteIP)

			if err := c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
				logger.HandlerAuth().Error("Failed to redirect after successful login from %s: %v", remoteIP, err)
				return echo.NewHTTPError(http.StatusInternalServerError,
					"failed to redirect to profile page")
			}
			return nil
		} else {
			logger.HandlerAuth().Warn("Failed login attempt from %s", remoteIP)
		}
	}

	page := loginpage.Page(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerAuth().Error("Failed to render login page for %s: %v", remoteIP, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render login page: "+err.Error())
	}

	return nil
}

// handleLogout handles user logout.
func (h *Auth) handleLogout(c echo.Context) error {
	remoteIP := c.RealIP()

	// Try to get user info before logout for better logging
	var userName string = "unknown"
	if user, err := helpers.GetUserFromContext(c); err == nil {
		userName = user.Name
		logger.HandlerAuth().Info("User %s logging out", user.Name)
	}

	if cookie, err := c.Cookie(constants.CookieName); err == nil {

		if err := h.DB.Cookies.Remove(cookie.Value); err != nil {
			logger.HandlerAuth().Error("Failed to remove cookie from database for user %s from %s: %v",
				userName, remoteIP, err)
		}
	}

	if err := c.Redirect(http.StatusSeeOther, "./login"); err != nil {
		logger.HandlerAuth().Error("Failed to redirect to login page after logout from %s: %v", remoteIP, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to redirect to login page")
	}

	return nil
}

// processApiKeyLogin processes API key authentication and creates a session.
func (h *Auth) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	remoteIP := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	// Validate API key format
	if len(apiKey) < 16 {
		return false
	}

	user, err := h.DB.Users.GetUserFromApiKey(apiKey)

	if err != nil {
		if !errors.Is(err, dberror.ErrNotFound) {
			logger.HandlerAuth().Error("Database error during authentication from %s: %v", remoteIP, err)
		}
		return false
	}

	// Check for existing session
	if existingCookie, err := ctx.Cookie(constants.CookieName); err == nil {

		if err := h.DB.Cookies.Remove(existingCookie.Value); err != nil {
			logger.HandlerAuth().Error("Failed to remove existing cookie for user %s from %s: %v",
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
		logger.HandlerAuth().Error("Failed to create session for user %s from %s: %v",
			user.Name, remoteIP, err)
		return false
	}

	// Log security-relevant information
	if user.IsAdmin() {
		logger.HandlerAuth().Info("Administrator %s logged in from %s", user.Name, remoteIP)
	}

	return true
}
