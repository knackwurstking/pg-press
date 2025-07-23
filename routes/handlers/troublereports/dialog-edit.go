// TODO:
//   - Update this dialog to allow uploading attachments, currently with a 10mb limit per attachment
//   - Attachments order will be set per attachment ID value
//   - Before submitting, the user should be allowed to reorder attachments manually
package troublereports

import (
	"net/http"
	"net/url"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type EditDialogTemplateData struct {
	ID                int                 `json:"id"`
	Submitted         bool                `json:"submitted"`
	Title             string              `json:"title"`
	Content           string              `json:"content"`
	LinkedAttachments []*pgvis.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                `json:"invalid_title"`
	InvalidContent    bool                `json:"invalid_content"`
}

func (h *Handler) handleGetDialogEdit(c echo.Context, pageData *EditDialogTemplateData) *echo.HTTPError {
	if pageData == nil {
		pageData = &EditDialogTemplateData{}
	}

	if c.QueryParam(constants.QueryParamCancel) == constants.TrueValue {
		pageData.Submitted = true
	}

	if !pageData.Submitted && !pageData.InvalidTitle && !pageData.InvalidContent {
		if idStr := c.QueryParam(constants.QueryParamID); idStr != "" {
			id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
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
			constants.TroubleReportsDialogEditComponentTemplatePath,
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

	title, content, herr := h.validateDialogEditFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData.Title = title
	dialogEditData.Content = content
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		modified := pgvis.NewModified[pgvis.TroubleReportMod](user, pgvis.TroubleReportMod{
			Title:             title,
			Content:           content,
			LinkedAttachments: []*pgvis.Attachment{}, // NOTE: Not implemented yet
		})
		tr := pgvis.NewTroubleReport(title, content, modified)

		if err := h.db.TroubleReports.Add(tr); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *Handler) handlePutDialogEdit(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, herr := h.validateDialogEditFormData(c)
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

		tr := pgvis.NewTroubleReport(title, content, trOld.Mods...)
		tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
			Title:             tr.Title,
			Content:           tr.Content,
			LinkedAttachments: tr.LinkedAttachments,
		}))

		if err := h.db.TroubleReports.Update(id, tr); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *Handler) validateDialogEditFormData(ctx echo.Context) (title, content string, httpErr *echo.HTTPError) {
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
