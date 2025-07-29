package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/htmxhandler"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

// ProfilePageData contains the data structure passed to the profile page template.
type ProfileTemplateData struct {
	User    *database.User     `json:"user"`
	Cookies []*database.Cookie `json:"cookies"`
}

// CookiesSorted returns the user's cookies sorted by last login time.
func (p *ProfileTemplateData) CookiesSorted() []*database.Cookie {
	return database.SortCookies(p.Cookies)
}

type Profile struct {
	*Base
}

func (h *Profile) RegisterRoutes(e *echo.Echo) {
	prefix := "/profile"

	e.GET(h.ServerPathPrefix+prefix, h.handleMainPage)

	htmxProfile := htmxhandler.Profile{Base: h.NewHTMX(prefix)}
	htmxProfile.RegisterRoutes(e)
}

func (h *Profile) handleMainPage(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	pageData := &ProfileTemplateData{
		User:    user,
		Cookies: make([]*database.Cookie, 0),
	}

	if err := h.handleUserNameChange(c, pageData); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	if cookies, err := h.DB.Cookies.ListApiKey(user.ApiKey); err == nil {
		pageData.Cookies = cookies
	}

	return utils.HandleTemplate(c, pageData, h.Templates,
		constants.ProfilePageTemplates)
}

func (h *Profile) handleUserNameChange(c echo.Context, pageData *ProfileTemplateData) error {
	formParams, _ := c.FormParams()
	userName := utils.SanitizeInput(formParams.Get(constants.UserNameFormField))

	if userName == "" || userName == pageData.User.UserName {
		return nil
	}

	if len(userName) < constants.UserNameMinLength || len(userName) > constants.UserNameMaxLength {
		return database.NewValidationError(constants.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	updatedUser := database.NewUser(pageData.User.TelegramID, userName,
		pageData.User.ApiKey)
	if err := h.DB.Users.Update(pageData.User.TelegramID, updatedUser); err != nil {
		return err
	}

	pageData.User.UserName = userName
	return nil
}
