package html

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models"
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
	formParams, _ := c.FormParams()
	apiKey := formParams.Get(constants.APIKeyFormField)

	if apiKey != "" && h.processApiKeyLogin(apiKey, c) {
		if err := c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to redirect to profile page")
		}
		return nil
	}

	page := loginpage.Page(apiKey, apiKey != "")
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render login page: "+err.Error())
	}
	return nil
}

// handleLogout handles user logout.
func (h *Auth) handleLogout(c echo.Context) error {
	if cookie, err := c.Cookie(constants.CookieName); err == nil {
		if err := h.DB.Cookies.Remove(cookie.Value); err != nil {
			logger.HandlerAuth().Error("Failed to remove cookie from database: %v", err)
		}
	}

	if err := c.Redirect(http.StatusSeeOther, "./login"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to redirect to login page")
	}

	return nil
}

// processApiKeyLogin processes API key authentication and creates a session.
func (h *Auth) processApiKeyLogin(apiKey string, ctx echo.Context) bool {
	user, err := h.DB.UsersHelper.GetUserFromApiKey(apiKey)
	if err != nil {
		if !errors.Is(err, dberror.ErrNotFound) {
			logger.HandlerAuth().Error("Failed to get user from API key: %v", err)
		}
		return false
	}

	if existingCookie, err := ctx.Cookie(constants.CookieName); err == nil {
		logger.HandlerAuth().Info("Removing existing authentication cookie")
		if err := h.DB.Cookies.Remove(existingCookie.Value); err != nil {
			logger.HandlerAuth().Error("Failed to remove existing cookie: %v", err)
		}
	}

	logger.HandlerAuth().Info("Creating new session for user %s (Telegram ID: %d)",
		user.UserName, user.TelegramID)

	cookieValue := uuid.New().String()
	cookie := &http.Cookie{
		Name:    constants.CookieName,
		Value:   cookieValue,
		Expires: time.Now().Add(constants.CookieExpirationDuration),
	}
	ctx.SetCookie(cookie)

	sessionCookie := models.NewCookie(
		ctx.Request().UserAgent(), cookieValue, apiKey)
	if err := h.DB.Cookies.Add(sessionCookie); err != nil {
		logger.HandlerAuth().Error("Failed to create session: %v", err)
		return false
	}

	return true
}
