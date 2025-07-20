// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

const (
	adminPrivilegesRequiredMessage = "administrator privileges required"
	invalidContentFormFieldMessage = "invalid content form value"
	invalidTitleFormFieldMessage   = "invalid title form value"
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
	e.GET(h.serverPathPrefix+"/trouble-reports", h.handleMainPage)

	editDialogPath := h.serverPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)

	dataPath := h.serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	modificationsPath := h.serverPathPrefix + "/trouble-reports/modifications"
	e.GET(modificationsPath, h.handleGetModifications)
}

func (h *Handler) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.templates,
		constants.TroubleReportsPageTemplates,
	)
}

func (h *Handler) handleGetDialogEdit(c echo.Context, pageData *EditDialogTemplateData) error {
	if pageData == nil {
		pageData = &EditDialogTemplateData{}
	}

	if c.QueryParam(constants.CancelQueryParam) == constants.TrueValue {
		pageData.Submitted = true
	}

	if !pageData.Submitted && !pageData.InvalidTitle && !pageData.InvalidContent {
		if idStr := c.QueryParam(constants.IDQueryParam); idStr != "" {
			id, herr := utils.ParseIDQuery(c, constants.IDQueryParam)
			if herr != nil {
				return herr
			}

			pageData.ID = int(id)

			tr, err := h.db.TroubleReports.Get(id)
			if err != nil {
				return utils.HandlePgvisError(c, err)
			}

			pageData.Title = tr.Title
			pageData.Content = tr.Content
			pageData.LinkedAttachments = tr.LinkedAttachments
		}
	}

	return utils.HandleTemplate(c, pageData,
		h.templates,
		[]string{
			constants.LegacyTroubleReportsDialogTemplatePath,
		},
	)
}

func (h *Handler) handlePostDialogEdit(c echo.Context) error {
	dialogEditData := &EditDialogTemplateData{
		Submitted: true,
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, herr := h.validateFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData.Title = title
	dialogEditData.Content = content
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
		tr := pgvis.NewTroubleReport(modified, title, content)

		if err := h.db.TroubleReports.Add(tr); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *Handler) handlePutDialogEdit(c echo.Context) error {
	id, herr := utils.ParseIDQuery(c, constants.IDQueryParam)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, herr := h.validateFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData := &EditDialogTemplateData{
		Submitted:      true,
		ID:             int(id),
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		trOld, err := h.db.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}

		modified := pgvis.NewModified(user, trOld)
		trNew := pgvis.NewTroubleReport(modified, title, content)

		if err := h.db.TroubleReports.Update(id, trNew); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *Handler) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.db.TroubleReports.List()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		TroubleReportsTemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.templates,
		[]string{
			constants.LegacyTroubleReportsDataTemplatePath,
		},
	)
}

func (h *Handler) handleDeleteData(c echo.Context) error {
	id, herr := utils.ParseIDQuery(c, "id")
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if !user.IsAdmin() {
		return echo.NewHTTPError(
			http.StatusForbidden,
			adminPrivilegesRequiredMessage,
		)
	}

	log.Infof("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if err := h.db.TroubleReports.Remove(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}

func (h *Handler) handleGetModifications(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	return utils.HandleTemplate(
		c,
		ModificationsTemplateData{
			User: user,
		},
		h.templates,
		[]string{
			constants.LegacyTroubleReportsModificationsTemplatePath,
		},
	)
}

func (h *Handler) validateFormData(ctx echo.Context) (title, content string, httpErr *echo.HTTPError) {
	var err error

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, invalidTitleFormFieldMessage))
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, invalidContentFormFieldMessage))
	}
	content = utils.SanitizeInput(content)

	return title, content, nil
}

type EditDialogTemplateData struct {
	ID                int                 `json:"id"`
	Submitted         bool                `json:"submitted"`
	Title             string              `json:"title"`
	Content           string              `json:"content"`
	LinkedAttachments []*pgvis.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                `json:"invalid_title"`
	InvalidContent    bool                `json:"invalid_content"`
}

type TroubleReportsTemplateData struct {
	TroubleReports []*pgvis.TroubleReport `json:"trouble_reports"`
	User           *pgvis.User            `json:"user"`
}

type ModificationsTemplateData struct {
	User *pgvis.User
}
