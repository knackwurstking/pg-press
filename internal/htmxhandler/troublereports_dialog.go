package htmxhandler

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/utils"
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
				return utils.HandlepgpressError(c, err)
			}

			pageData.Title = tr.Title
			pageData.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReportsHelper.LoadAttachments(tr); err == nil {
				pageData.LinkedAttachments = loadedAttachments
			}
		}
	}

	// TODO: Migrate to templ
	return utils.HandleTemplate(c, pageData,
		h.Templates,
		[]string{
			//constants.HTMXTroubleReportsDialogEditTemplatePath,
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

		if err := h.DB.TroubleReportsHelper.AddWithAttachments(tr, attachments); err != nil {
			return utils.HandlepgpressError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

// FIXME: Attachment changes does not add a modified to mods it seems
func (h *TroubleReports) handlePutDialogEdit(c echo.Context) error {
	// Get ID from query parameter
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	// Get user from context
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Get Title, Content and Attachments from form data
	title, content, attachments, herr := h.validateDialogEditFormData(c)
	if herr != nil {
		return herr
	}

	// Initialize dialog template data
	dialogEditData := &dialogEditTemplateData{
		Submitted:      true,
		ID:             int(id),
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	// Abort if invalid title or content
	if dialogEditData.InvalidTitle || dialogEditData.InvalidContent {
		dialogEditData.Submitted = false
		return h.handleGetDialogEdit(c, dialogEditData)
	}

	// Set attachments to handlePutDialogEdit
	dialogEditData.LinkedAttachments = attachments

	// Query previous trouble report
	trOld, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlepgpressError(c, err)
	}

	// Create new trouble report
	tr := database.NewTroubleReport(title, content, trOld.Mods...)

	// Filter out existing and new attachments
	var existingAttachmentIDs []int64
	var newAttachments []*database.Attachment
	for _, a := range dialogEditData.LinkedAttachments {
		if a.GetID() > 0 {
			existingAttachmentIDs = append(existingAttachmentIDs, a.GetID())
		} else {
			newAttachments = append(newAttachments, a)
		}
	}

	// Update trouble report with existing and new attachments, title content and mods
	tr.LinkedAttachments = existingAttachmentIDs
	tr.Mods = append(tr.Mods, database.NewModified(user, database.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))
	if err := h.DB.TroubleReportsHelper.UpdateWithAttachments(id, tr, newAttachments); err != nil {
		return utils.HandlepgpressError(c, err)
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
