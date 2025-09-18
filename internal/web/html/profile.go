package html

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/profilepage"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
)

type Profile struct {
	*handlers.BaseHandler
}

func NewProfile(db *database.DB) *Profile {
	return &Profile{
		BaseHandler: handlers.NewBaseHandler(db, logger.HandlerProfile()),
	}
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/profile", h.HandleProfile),
		},
	)
}

func (h *Profile) HandleProfile(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("Rendering profile page for user %s", user.Name)

	if err = h.handleUserNameChange(c, user); err != nil {
		return h.HandleError(c, err, "error updating username")
	}

	page := profilepage.Page(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render profile page: "+err.Error())
	}
	return nil
}

func (h *Profile) handleUserNameChange(c echo.Context, user *models.User) error {
	userName := h.GetSanitizedFormValue(c, constants.UserNameFormField)

	if userName == "" || userName == user.Name {
		return nil
	}

	if len(userName) < UserNameMinLength || len(userName) > UserNameMaxLength {
		return utils.NewValidationError(constants.UserNameFormField + ": " +
			"username must be between 1 and 100 characters")
	}

	h.LogInfo("User %s (Telegram ID: %d) is changing username to %s",
		user.Name, user.TelegramID, userName)

	updatedUser := models.NewUser(user.TelegramID, userName, user.ApiKey)
	updatedUser.LastFeed = user.LastFeed

	if err := h.DB.Users.Update(updatedUser, user); err != nil {
		return err
	}

	h.LogInfo("Successfully updated username for user %d from %s to %s",
		user.TelegramID, user.Name, userName)

	user.Name = userName

	return nil
}
