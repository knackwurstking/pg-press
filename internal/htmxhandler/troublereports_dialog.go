package htmxhandler

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

func (h *TroubleReports) handleGetDialogEdit(
	c echo.Context,
	pageData *dialogEditTemplateData,
) *echo.HTTPError {
	if pageData == nil {
		pageData = &dialogEditTemplateData{}
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

			tr, err := h.DB.TroubleReports.Get(id)
			if err != nil {
				return utils.HandlePgvisError(c, err)
			}

			pageData.Title = tr.Title
			pageData.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReportService.LoadAttachments(tr); err == nil {
				pageData.LinkedAttachments = loadedAttachments
			}
		}
	}

	return utils.HandleTemplate(c, pageData,
		h.Templates,
		[]string{
			constants.HTMXTroubleReportsDialogEditTemplatePath,
		},
	)
}

func (h *TroubleReports) handlePostDialogEdit(c echo.Context) error {
	dialogEditData := &dialogEditTemplateData{
		Submitted: true,
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, attachments, herr := h.validateDialogEditFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData.Title = title
	dialogEditData.Content = content
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		dialogEditData.LinkedAttachments = attachments
		modified := database.NewModified[database.TroubleReportMod](user, database.TroubleReportMod{
			Title:             title,
			Content:           content,
			LinkedAttachments: []int64{}, // Will be set by the service
		})
		tr := database.NewTroubleReport(title, content, modified)

		if err := h.DB.TroubleReportService.AddWithAttachments(tr, attachments); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *TroubleReports) handlePutDialogEdit(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, attachments, herr := h.validateDialogEditFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData := &dialogEditTemplateData{
		Submitted:      true,
		ID:             int(id),
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		dialogEditData.LinkedAttachments = attachments
		trOld, err := h.DB.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}

		tr := database.NewTroubleReport(title, content, trOld.Mods...)

		// Convert existing attachments to IDs for reordering
		var existingAttachmentIDs []int64
		for _, att := range attachments {
			if att.GetID() > 0 {
				existingAttachmentIDs = append(existingAttachmentIDs, att.GetID())
			}
		}

		// Filter out new attachments
		var newAttachments []*database.Attachment
		for _, att := range attachments {
			if att.GetID() == 0 {
				newAttachments = append(newAttachments, att)
			}
		}

		tr.LinkedAttachments = existingAttachmentIDs
		tr.Mods = append(tr.Mods, database.NewModified(user, database.TroubleReportMod{
			Title:             tr.Title,
			Content:           tr.Content,
			LinkedAttachments: []int64{},
		}))

		if err := h.DB.TroubleReportService.UpdateWithAttachments(id, tr, newAttachments); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *TroubleReports) validateDialogEditFormData(ctx echo.Context) (
	title, content string,
	attachments []*database.Attachment,
	httpErr *echo.HTTPError,
) {
	var err error

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, invalidTitleFormFieldMessage))
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, invalidContentFormFieldMessage))
	}
	content = utils.SanitizeInput(content)

	// Process existing attachments and their order
	attachments, err = h.processAttachments(ctx)
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, "failed to process attachments"))
	}

	return title, content, attachments, nil
}
