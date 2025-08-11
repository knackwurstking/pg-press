package htmxhandler

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

func (h *TroubleReports) handleGetDialogEdit(
	c echo.Context,
	props *components.TroubleReportsEditDialogProps,
) *echo.HTTPError {
	if props == nil {
		props = &components.TroubleReportsEditDialogProps{}
	}

	if c.QueryParam(constants.QueryParamCancel) == constants.TrueValue {
		props.Submitted = true
	}

	if !props.Submitted && !props.InvalidTitle && !props.InvalidContent {
		if idStr := c.QueryParam(constants.QueryParamID); idStr != "" {
			id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
			if herr != nil {
				return herr
			}

			props.ID = id

			tr, err := h.DB.TroubleReports.Get(id)
			if err != nil {
				return utils.HandlepgpressError(c, err)
			}

			props.Title = tr.Title
			props.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReportsHelper.LoadAttachments(tr); err == nil {
				props.Attachments = loadedAttachments
			}
		}
	}

	dialog := components.TroubleReportsEditDialog(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render Trouble Reports Edit Dialog: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handlePostDialogEdit(c echo.Context) error {
	props := &components.TroubleReportsEditDialogProps{
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

	props.Title = title
	props.Content = content
	props.InvalidTitle = title == ""
	props.InvalidContent = content == ""

	if !props.InvalidTitle && !props.InvalidContent {
		props.Attachments = attachments
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
		props.Submitted = false
	}

	return h.handleGetDialogEdit(c, props)
}

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
	props := &components.TroubleReportsEditDialogProps{
		Submitted:      true,
		ID:             id,
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	// Abort if invalid title or content
	if props.InvalidTitle || props.InvalidContent {
		props.Submitted = false
		return h.handleGetDialogEdit(c, props)
	}

	// Set attachments to handlePutDialogEdit
	props.Attachments = attachments

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
	for _, a := range props.Attachments {
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

	return h.handleGetDialogEdit(c, props)
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
