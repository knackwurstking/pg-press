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

	attachmentID := c.QueryParam("attachment_id")
	if attachmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	// Get trouble report
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find the attachment
	var attachment *pgvis.Attachment
	for _, att := range tr.LinkedAttachments {
		if att.ID == attachmentID {
			attachment = att
			break
		}
	}

	if attachment == nil {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// Set appropriate headers
	c.Response().Header().Set("Content-Type", attachment.MimeType)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(attachment.Data)))

	// Try to determine filename from attachment ID
	filename := attachment.ID
	if ext := attachment.GetFileExtension(); ext != "" {
		if !strings.HasSuffix(filename, ext) {
			filename += ext
		}
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

	// Reorder attachments based on the new order
	reorderedAttachments := make([]*pgvis.Attachment, 0, len(tr.LinkedAttachments))
	for _, attachmentID := range orderParts {
		attachmentID = strings.TrimSpace(attachmentID)
		if attachmentID == "" {
			continue
		}

		// Find the attachment with this ID
		for _, attachment := range tr.LinkedAttachments {
			if attachment.ID == attachmentID {
				reorderedAttachments = append(reorderedAttachments, attachment)
				break
			}
		}
	}

	// Update the trouble report with reordered attachments
	tr.LinkedAttachments = reorderedAttachments
	tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return updated dialog
	pageData := &DialogEditTemplateData{
		Submitted:         true, // Prevent reloading from database
		ID:                int(id),
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}

	return h.dialogEditHandler.handleGetDialogEdit(c, pageData)
}

// handleDeleteAttachment handles deletion of a specific attachment
func (h *AttachmentsHandler) handleDeleteAttachment(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentID := c.QueryParam("attachment_id")
	if attachmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
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

	// Find and remove the attachment
	newAttachments := make([]*pgvis.Attachment, 0, len(tr.LinkedAttachments))
	found := false

	// Trim whitespace from attachment ID to avoid comparison issues
	attachmentID = strings.TrimSpace(attachmentID)

	for _, attachment := range tr.LinkedAttachments {
		trimmedAttachmentID := strings.TrimSpace(attachment.ID)
		if trimmedAttachmentID != attachmentID {
			newAttachments = append(newAttachments, attachment)
		} else {
			found = true
		}
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// Update the trouble report
	tr.LinkedAttachments = newAttachments
	tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return only the attachments section HTML
	return h.renderAttachmentsSection(c, int(id), tr.LinkedAttachments)
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
