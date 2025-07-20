// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"

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
	e.GET(h.serverPathPrefix+"/trouble-reports", h.handleMainPage)

	editDialogPath := h.serverPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetEditDialog(c, nil)
	})
	e.POST(editDialogPath, func(c echo.Context) error {
		return h.handlePostDialogEdit(c)
	})
	e.PUT(editDialogPath, handleUpdateReport(h.templates, h.db)) // TODO: ...

	dataPath := h.serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, handleGetData(h.templates, h.db))         // TODO: ...
	e.DELETE(dataPath, handleDeleteReport(h.templates, h.db)) // TODO: ...

	modificationsPath := h.serverPathPrefix + "/trouble-reports/modifications"
	e.GET(modificationsPath, handleGetModifications(h.templates, h.db)) // TODO: ...
}

func (h *Handler) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.templates,
		constants.TroubleReportsPageTemplates,
	)
}

func (h *Handler) handleGetEditDialog(c echo.Context, pageData *EditDialogTemplateData) error {
	if pageData == nil {
		pageData = &EditDialogTemplateData{}
	}

	if c.QueryParam(constants.CancelQueryParam) == constants.TrueValue {
		pageData.Submitted = true
	}

	if !pageData.Submitted && !pageData.InvalidTitle && !pageData.InvalidContent {
		if idStr := c.QueryParam(constants.IDQueryParam); idStr != "" {
			id, herr := utils.ParseRequiredIDQuery(c, constants.IDQueryParam)
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

	title, content, herr := extractAndValidateFormData(c)
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

	return h.handleGetEditDialog(c, dialogEditData)
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

// .................................................................................... //

func handleUpdateReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return PUTDialogEdit(templates, c, db)
	}
}

func handleGetData(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETData(templates, c, db)
	}
}

func handleDeleteReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return DELETEData(templates, c, db)
	}
}

func handleGetModifications(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETModifications(templates, c, db)
	}
}
