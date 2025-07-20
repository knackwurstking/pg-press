// Package profile provides HTTP route handlers for user profile management.
package profile

import (
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/profile", h.handleMainPage)
	e.GET(h.serverPathPrefix+"/profile/cookies", h.handleGetCookies)
	e.DELETE(h.serverPathPrefix+"/profile/cookies", h.handleDeleteCookies)
}

// Profile contains the data structure passed to the profile page template.
type ProfilePageData struct {
	User    *pgvis.User     `json:"user"`
	Cookies []*pgvis.Cookie `json:"cookies"`
}

// CookiesSorted returns the user's cookies sorted by last login time.
func (p *ProfilePageData) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

func (h *Handler) handleMainPage(c echo.Context) error {
	pageData := &ProfilePageData{
		Cookies: make([]*pgvis.Cookie, 0),
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}
	pageData.User = user

	if err := h.handleUserNameChange(c, pageData); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	if cookies, err := h.db.Cookies.ListApiKey(pageData.User.ApiKey); err == nil {
		pageData.Cookies = cookies
	}

	return utils.HandleTemplate(c, pageData,
		h.templates,
		constants.ProfilePageTemplates,
	)
}

func (h *Handler) handleGetCookies(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	cookies, err := h.db.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	cookies = pgvis.SortCookies(cookies)

	return utils.HandleTemplate(c, cookies,
		h.templates,
		[]string{
			constants.ProfileCookiesComponentTemplatePath,
		},
	)
}

func (h *Handler) handleDeleteCookies(c echo.Context) error {
	value := utils.SanitizeInput(c.QueryParam("value"))
	if value == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"cookie value parameter is required",
		)
	}

	if err := h.db.Cookies.Remove(value); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetCookies(c)
}

func (h *Handler) handleUserNameChange(c echo.Context, pageData *ProfilePageData) error {
	formParams, _ := c.FormParams()
	userName := utils.SanitizeInput(formParams.Get(constants.UserNameFormField))

	if userName == "" || userName == pageData.User.UserName {
		return nil
	}

	if len(userName) < constants.UserNameMinLength || len(userName) > constants.UserNameMaxLength {
		return pgvis.NewValidationError(constants.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	updatedUser := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)

	if err := h.db.Users.Update(pageData.User.TelegramID, updatedUser); err != nil {
		return err
	}

	pageData.User.UserName = userName
	return nil
}
