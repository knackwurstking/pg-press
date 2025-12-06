package profile

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/profile/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
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

func (h *Handler) GetProfilePage(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = h.handleUserNameChange(c, user)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Profile Page")
	}

	return nil
}

func (h *Handler) HTMXGetCookies(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookies, merr := h.registry.Cookies.ListApiKey(user.ApiKey)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Cookies(models.SortCookies(cookies))
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cookies")
	}

	return nil
}

func (h *Handler) HTMXDeleteCookies(c echo.Context) error {
	slog.Info("Eat the cookie from a user :)")

	value, merr := utils.ParseQueryString(c, "value")
	if merr != nil {
		return merr.Echo()
	}

	merr = h.registry.Cookies.Remove(value)
	if merr != nil {
		return merr.Echo()
	}

	err := h.HTMXGetCookies(c)
	if err != nil {
		return errors.NewMasterError(err, 0).Echo()
	}

	return nil
}

func (h *Handler) handleUserNameChange(c echo.Context, user *models.User) *errors.MasterError {
	userName := c.FormValue("user-name")
	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < env.UserNameMinLength || len(userName) > env.UserNameMaxLength {
		return errors.NewMasterError(
			fmt.Errorf("username must be between 1 and 100 characters"),
			http.StatusBadRequest,
		)
	}

	slog.Info("Changing user name", "user_from", user.Name, "user_to", userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	merr := h.registry.Users.Update(updatedUser)
	if merr != nil {
		return merr
	}

	// Create feed entry
	feedTitle := "Benutzername geändert"
	feedContent := fmt.Sprintf("Alter Name: %s\nNeuer Name: %s", user.Name, userName)
	_, merr = h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for username change", "error", merr)
	}

	user.Name = userName

	return nil
}
