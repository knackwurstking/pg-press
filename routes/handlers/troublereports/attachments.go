package troublereports

import (
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type AttachmentsHandler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS

	dialogEditHandler *DialogEditHandler
}

func (h *AttachmentsHandler) RegisterRoutes(e *echo.Echo) {
	attachmentReorderPath := h.serverPathPrefix + "/trouble-reports/attachments/reorder"

	e.POST(attachmentReorderPath, h.handlePostAttachmentReorder)

	attachmentPath := h.serverPathPrefix + "/trouble-reports/attachments"

	e.GET(attachmentPath, h.handleGetAttachment)
	e.DELETE(attachmentPath, h.handleDeleteAttachment)
}

// handleGetAttachment handles downloading/viewing of attachments
func (h *AttachmentsHandler) handleGetAttachment(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentIDStr := c.QueryParam("attachment_id")
	if attachmentIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment_id parameter")
	}

	// Get trouble report to verify attachment belongs to it
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Check if attachment ID is in the trouble report's linked attachments
	var found bool
	for _, linkedID := range tr.LinkedAttachments {
		if linkedID == attachmentID {
			found = true
			break
		}
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found in this trouble report")
	}

	// Get the attachment from the attachments table
	attachment, err := h.db.Attachments.Get(attachmentID)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Set appropriate headers
	c.Response().Header().Set("Content-Type", attachment.MimeType)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(attachment.Data)))

	// Try to determine filename from attachment ID
	filename := fmt.Sprintf("attachment_%d", attachmentID)
	if ext := attachment.GetFileExtension(); ext != "" {
		filename += ext
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}

// handlePostAttachmentReorder handles reordering of attachments
func (h *AttachmentsHandler) handlePostAttachmentReorder(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Get the new order from form data
	newOrder := c.FormValue("new_order")
	if newOrder == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing new_order parameter")
	}

	// Parse the new order (comma-separated attachment IDs)
	orderParts := strings.Split(newOrder, ",")

	// Get existing trouble report
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Reorder attachment IDs based on the new order
	reorderedAttachmentIDs := make([]int64, 0, len(tr.LinkedAttachments))
	for _, attachmentIDStr := range orderParts {
		attachmentIDStr = strings.TrimSpace(attachmentIDStr)
		if attachmentIDStr == "" {
			continue
		}

		attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
		if err != nil {
			continue // Skip invalid IDs
		}

		// Check if this ID exists in the current attachments
		for _, existingID := range tr.LinkedAttachments {
			if existingID == attachmentID {
				reorderedAttachmentIDs = append(reorderedAttachmentIDs, attachmentID)
				break
			}
		}
	}

	// Update the trouble report with reordered attachment IDs
	tr.LinkedAttachments = reorderedAttachmentIDs
	tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Load attachments for dialog
	attachments, err := h.db.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return updated dialog
	pageData := &DialogEditTemplateData{
		Submitted:         true, // Prevent reloading from database
		ID:                int(id),
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: attachments,
	}

	return h.dialogEditHandler.handleGetDialogEdit(c, pageData)
}

// handleDeleteAttachment handles deletion of a specific attachment
func (h *AttachmentsHandler) handleDeleteAttachment(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentIDStr := c.QueryParam("attachment_id")
	if attachmentIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment_id parameter")
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Get existing trouble report
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find and remove the attachment ID
	newAttachmentIDs := make([]int64, 0, len(tr.LinkedAttachments))
	found := false

	for _, linkedID := range tr.LinkedAttachments {
		if linkedID != attachmentID {
			newAttachmentIDs = append(newAttachmentIDs, linkedID)
		} else {
			found = true
		}
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// Update the trouble report
	tr.LinkedAttachments = newAttachmentIDs
	tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Load attachments and return only the attachments section HTML
	attachments, err := h.db.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.renderAttachmentsSection(c, int(id), attachments)
}

// renderAttachmentsSection renders only the attachments section HTML
func (h *AttachmentsHandler) renderAttachmentsSection(c echo.Context, reportID int, attachments []*pgvis.Attachment) error {
	data := struct {
		ID                int                 `json:"id"`
		LinkedAttachments []*pgvis.Attachment `json:"linked_attachments"`
	}{
		ID:                reportID,
		LinkedAttachments: attachments,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.templates,
		[]string{constants.AttachmentsSectionComponentTemplatePath},
	)
}
