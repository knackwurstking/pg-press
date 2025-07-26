package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/logger"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

type AuthTemplateData struct {
	ApiKey        string
	InvalidApiKey bool
}

type Auth struct {
	*Base
}

// NewHandler creates a new authentication handler.
func NewAuth(base *Base) *Auth {
	return &Auth{base}
}

func (h *Auth) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/login", h.handleLogin)
	e.GET(h.ServerPathPrefix+"/logout", h.handleLogout)
}

// handleLogin handles the login page and form submission.
func (h *Auth) handleLogin(c echo.Context) error {
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
		AuthTemplateData{
			ApiKey:        apiKey,
			InvalidApiKey: apiKey != "",
		},
		h.Templates,
		constants.LoginPageTemplates,
	)
}

// handleLogout handles user logout.
func (h *Auth) handleLogout(c echo.Context) error {
	if cookie, err := c.Cookie(constants.CookieName); err == nil {
		if err := h.DB.Cookies.Remove(cookie.Value); err != nil {
			logger.Auth().Error("Failed to remove cookie from database: %v", err)
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
func (h *Auth) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	if apiKey == "" {
		return false
	}

	user, err := h.DB.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false
		}

		logger.Auth().Error("Failed to get user from API key: %v", err)
		return false
	}

	if user.ApiKey != apiKey {
		return false
	}

	if existingCookie, err := ctx.Cookie(constants.CookieName); err == nil {
		logger.Auth().Info("Removing existing authentication cookie")

		if err := h.DB.Cookies.Remove(existingCookie.Value); err != nil {
			logger.Auth().Error("Failed to remove existing cookie: %v", err)
		}
	}

	logger.Auth().Info("Creating new session for user %s (Telegram ID: %d)",
		user.UserName, user.TelegramID)

	cookie := &http.Cookie{
		Name:    constants.CookieName,
		Value:   uuid.New().String(),
		Expires: time.Now().Add(constants.CookieExpirationDuration),
	}

	ctx.SetCookie(cookie)

	sessionCookie := database.NewCookie(
		ctx.Request().UserAgent(),
		cookie.Value,
		apiKey,
	)

	if err := h.DB.Cookies.Add(sessionCookie); err != nil {
		logger.Auth().Error("Failed to create session: %v", err)
		return false
	}

	return true
}
