package htmxhandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/utils"
)

func (h *TroubleReports) handleGetAttachment(c echo.Context) error {
	attachmentID, err := utils.ParseInt64Query(c, constants.QueryParamAttachmentID)
	if err != nil {
		return err
	}

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get attachment: "+err.Error())
	}

	// Set appropriate headers
	c.Response().Header().Set("Content-Type", attachment.MimeType)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(attachment.Data)))

	// Try to determine filename from attachment ID
	filename := fmt.Sprintf("attachment_%d", attachmentID)
	if ext := attachment.GetFileExtension(); ext != "" {
		filename += ext
	}
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}
