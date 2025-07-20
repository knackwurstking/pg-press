// Package auth provides HTTP route handlers for authentication.
package auth

import (
	"embed"
	"errors"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

// Handler handles authentication-related HTTP requests.
type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        embed.FS
}

// NewHandler creates a new authentication handler.
func NewHandler(db *pgvis.DB, serverPathPrefix string, templates embed.FS) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
	}
}

// RegisterRoutes registers all authentication routes.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/login", h.handleLogin)
	e.GET(h.serverPathPrefix+"/logout", h.handleLogout)
}

type LoginTemplateData struct {
	ApiKey        string
	InvalidApiKey bool
}

// handleLogin handles the login page and form submission.
func (h *Handler) handleLogin(c echo.Context) error {
	formParams, _ := c.FormParams()
	apiKey := formParams.Get(constants.APIKeyFormField)

	if apiKey != "" {
		if h.processApiKeyLogin(apiKey, c) {
			if err := c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					constants.RedirectFailedMessage,
				)
			}
			return nil
		}
	}

	return utils.HandleTemplate(
		c,
		LoginTemplateData{
			ApiKey:        apiKey,
			InvalidApiKey: apiKey != "",
		},
		h.templates,
		constants.LoginPageTemplates,
	)
}

// handleLogout handles user logout.
func (h *Handler) handleLogout(c echo.Context) error {
	if cookie, err := c.Cookie(constants.CookieName); err == nil {
		if err := h.db.Cookies.Remove(cookie.Value); err != nil {
			log.Errorf("Failed to remove cookie from database: %s", err)
		}
	}

	if err := c.Redirect(http.StatusSeeOther, "./login"); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			constants.RedirectFailedMessage,
		)
	}

	return nil
}

// processApiKeyLogin processes API key authentication and creates a session.
func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	if apiKey == "" {
		return false
	}

	user, err := h.db.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if errors.Is(err, pgvis.ErrNotFound) {
			return false
		}

		log.Errorf("Failed to get user from API key: %s", err)
		return false
	}

	if user.ApiKey != apiKey {
		return false
	}

	if existingCookie, err := ctx.Cookie(constants.CookieName); err == nil {
		log.Debug("Removing existing authentication cookie")

		if err := h.db.Cookies.Remove(existingCookie.Value); err != nil {
			log.Warnf("Failed to remove existing cookie: %s", err)
		}
	}

	log.Debugf("Creating new session for user %s (Telegram ID: %d)",
		user.UserName, user.TelegramID)

	cookie := &http.Cookie{
		Name:    constants.CookieName,
		Value:   uuid.New().String(),
		Expires: time.Now().Add(constants.CookieExpirationDuration),
	}

	ctx.SetCookie(cookie)

	sessionCookie := pgvis.NewCookie(
		ctx.Request().UserAgent(),
		cookie.Value,
		apiKey,
	)

	if err := h.db.Cookies.Add(sessionCookie); err != nil {
		log.Errorf("Failed to create session: %s", err)
		return false
	}

	return true
}
