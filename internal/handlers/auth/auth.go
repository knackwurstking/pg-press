package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/auth/templates"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgpress-api-key"
)

type Handler struct {
	db *common.DB
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path+"/login", h.GetLoginPage),
		ui.NewEchoRoute(http.MethodPost, path+"/login", h.PostLoginPage),
		ui.NewEchoRoute(http.MethodGet, path+"/logout", h.GetLogout),
	})
}

func (h *Handler) GetLoginPage(c echo.Context) *echo.HTTPError {
	var (
		invalid = shared.ParseQueryBool(c, "invalid")
		apiKey  = c.FormValue("api-key")
		page    = templates.LoginPage(apiKey, invalid)
	)

	err := page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Login Page")
	}

	return nil
}

func (h *Handler) PostLoginPage(c echo.Context) *echo.HTTPError {
	slog.Info("User authentication request received")

	// Parse form
	apiKey := c.FormValue("api-key")
	err := h.processApiKeyLogin(apiKey, c)
	if err != nil {
		slog.Warn("Processing api key failed", "api_key", shared.MaskString(apiKey), "error", err)
	}
	if apiKey == "" || err != nil {
		invalid := true
		merr := urlb.RedirectTo(c, urlb.UrlLogin(apiKey, &invalid).Page)
		if merr != nil {
			return merr.Echo()
		}
	}

	merr := urlb.RedirectTo(c, urlb.UrlProfile("").Page)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func (h *Handler) GetLogout(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookie, err := c.Cookie(CookieName)
	if err == nil {
		merr := h.db.User.Cookie.Delete(cookie.Value)
		if merr != nil {
			slog.Error("Failed to remove cookie", "user_name", user.Name, "error", merr)
		}
	}

	merr = urlb.RedirectTo(c, urlb.UrlLogin("", nil).Page)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func (h *Handler) processApiKeyLogin(apiKey string, ctx echo.Context) *errors.MasterError {
	if len(apiKey) < shared.MinAPIKeyLength {
		return errors.NewMasterError(fmt.Errorf("api key too short"), http.StatusBadRequest)
	}

	user, merr := helper.GetUserForApiKey(h.db, apiKey)
	if merr != nil {
		return merr
	}

	merr = h.clearExistingSession(ctx)
	if merr != nil {
		slog.Warn("Failed to clear existing session", "error", merr)
	}

	merr = h.createSession(ctx, user.ID)
	if merr != nil {
		return merr
	}

	if user.IsAdmin() {
		slog.Info("Administrator login successful", "user", user.Name, "ip", ctx.RealIP())
	}

	return nil
}

func (h *Handler) clearExistingSession(ctx echo.Context) *errors.MasterError {
	cookie, err := ctx.Cookie(CookieName)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	merr := h.db.User.Cookie.Delete(cookie.Value)
	if merr != nil {
		return merr
	}

	return nil
}

func (h *Handler) createSession(ctx echo.Context, userID shared.TelegramID) *errors.MasterError {
	cookie := &shared.Cookie{
		UserAgent: ctx.Request().UserAgent(),
		Value:     uuid.New().String(),
		UserID:    userID,
		LastLogin: shared.NewUnixMilli(time.Now()),
	}
	merr := h.db.User.Cookie.Create(cookie)
	if merr != nil {
		return merr
	}

	ctx.SetCookie(&http.Cookie{
		Name:     CookieName,
		Value:    cookie.Value,
		Expires:  cookie.ExpiredAtTime(),
		Path:     env.ServerPathPrefix + "/",
		HttpOnly: true,
		Secure:   ctx.Request().TLS != nil || ctx.Scheme() == "https",
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}
