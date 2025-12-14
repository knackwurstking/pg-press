package profile

import (
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/profile/templates"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB     *common.DB
	Logger *log.Logger
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		DB:     db,
		Logger: env.NewLogger("handler: profile: "),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			ui.NewEchoRoute(http.MethodGet, path, h.GetProfilePage),
			ui.NewEchoRoute(http.MethodGet, path+"/cookies", h.HTMXGetCookies),
			ui.NewEchoRoute(http.MethodDelete, path+"/cookies",
				h.HTMXDeleteCookies),
		},
	)
}

func (h *Handler) GetProfilePage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = h.handleUserNameChange(c, user)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(templates.PageProps{User: user})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Profile Page")
	}

	return nil
}

func (h *Handler) HTMXGetCookies(c echo.Context) *echo.HTTPError {
	return h.renderCookies(c, false)
}

func (h *Handler) HTMXDeleteCookies(c echo.Context) *echo.HTTPError {
	value, merr := shared.ParseQueryString(c, "value")
	if merr != nil {
		return merr.Echo()
	}

	merr = h.DB.User.Cookie.Delete(value)
	if merr != nil {
		return merr.Echo()
	}

	eerr := h.HTMXGetCookies(c)
	if eerr != nil {
		return eerr
	}

	return h.renderCookies(c, true)
}

func (h *Handler) handleUserNameChange(c echo.Context, user *shared.User) *errors.MasterError {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < shared.UserNameMinLength || len(userName) > shared.UserNameMaxLength {
		return errors.NewMasterError(
			fmt.Errorf("username must be between 1 and 100 characters"),
			http.StatusBadRequest,
		)
	}

	user.Name = userName
	merr := h.DB.User.User.Update(user)
	if merr != nil {
		return merr
	}
	h.Logger.Printf("User %s changed their name to %s\n", user.Name, userName)

	return nil
}

func (h *Handler) renderCookies(c echo.Context, oob bool) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookies, merr := helper.ListCookiesForApiKey(h.DB, user.ApiKey)
	if merr != nil {
		return merr.Echo()
	}

	// Sort cookies by last login
	slices.SortFunc(cookies, func(a, b *shared.Cookie) int {
		return int(a.LastLogin - b.LastLogin)
	})

	t := templates.Cookies(templates.CookiesProps{
		Cookies: cookies,
		OOB:     oob,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cookies")
	}
	return nil
}
