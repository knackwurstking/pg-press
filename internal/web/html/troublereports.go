package html

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/modificationspage"
	"github.com/knackwurstking/pgpress/internal/web/templates/troublereportspage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/modification"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports",
				h.handleTroubleReportsGET),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/share-pdf",
				h.handleSharePdfGET),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/attachment",
				h.handleAttachmentGET),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/modifications/:id",
				h.handleModificationsGET),

			//helpers.NewEchoRoute(http.MethodPost, "/trouble-reports/rollback/:id",
			//	h.handleRollbackPOST),
		},
	)
}

func (h *TroubleReports) handleTroubleReportsGET(c echo.Context) error {
	logger.HandlerTroubleReports().Debug("Rendering trouble reports page")

	page := troublereportspage.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble reports page: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleSharePdfGET(c echo.Context) error {
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	logger.HandlerTroubleReports().Info("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to retrieve trouble report: "+err.Error())
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Fehler beim Erstellen der PDF",
		)
	}

	logger.HandlerTroubleReports().Info(
		"Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, pdfBuffer.Len())

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *TroubleReports) handleAttachmentGET(c echo.Context) error {
	attachmentID, err := helpers.ParseInt64Query(c, "attachment_id")
	if err != nil {
		logger.HandlerTroubleReports().Error("Invalid attachment ID parameter: %v", err)
		return err
	}

	logger.HandlerTroubleReports().Debug("Fetching attachment %d", attachmentID)

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		logger.HandlerTroubleReports().Error("Failed to get attachment %d: %v", attachmentID, err)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
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

	logger.HandlerTroubleReports().Info("Serving attachment %d (size: %d bytes, type: %s)",
		attachmentID, len(attachment.Data), attachment.MimeType)

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}

func (h *TroubleReports) handleModificationsGET(c echo.Context) error {
	logger.HandlerTroubleReports().Info("Handling modifications for trouble report")

	// Parse ID parameter
	id, err := helpers.ParseInt64Param(c, "id")
	if err != nil {
		return err
	}

	// Fetch modifications for this trouble report
	logger.HandlerTroubleReports().Debug("Fetching modifications for trouble report %d", id)
	modifications, err := h.DB.Modifications.ListWithUser(
		services.ModificationTypeTroubleReport, id, 100, 0)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to retrieve modifications: "+err.Error())
	}

	logger.HandlerTroubleReports().Debug(
		"Found %d modifications for trouble report %d",
		len(modifications), id)

	// Convert to the format expected by the modifications page
	var m modification.Mods[models.TroubleReportModData]
	for _, mod := range modifications {
		var data models.TroubleReportModData
		if err := json.Unmarshal(mod.Modification.Data, &data); err != nil {
			continue
		}

		modEntry := modification.NewMod(&mod.User, data)
		modEntry.Time = mod.Modification.CreatedAt.UnixMilli()
		m = append(m, modEntry)
	}

	// Get user from context to check permissions
	currentUser, _ := helpers.GetUserFromContext(c)
	canRollback := currentUser != nil && currentUser.IsAdmin()

	// Create render function using the new template
	f := troublereportspage.CreateModificationRenderer(id, canRollback)

	// Rendering the page template
	page := modificationspage.Page(m, f)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble reports page: "+err.Error())
	}

	return nil
}

func (h *TroubleReports) shareResponse(
	c echo.Context,
	tr *models.TroubleReportWithAttachments,
	buf *bytes.Buffer,
) error {
	filename := fmt.Sprintf("fehlerbericht_%d_%s.pdf",
		tr.ID, time.Now().Format("2006-01-02"))

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Length",
		fmt.Sprintf("%d", buf.Len()))

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}

//func (h *TroubleReports) handleRollbackPOST(c echo.Context) error {
//	logger.HandlerTroubleReports().Info("Handling rollback for trouble report")
//
//	// Parse ID parameter
//	id, err := helpers.ParseInt64Param(c, "id")
//	if err != nil {
//		logger.HandlerTroubleReports().Error("Invalid ID parameter: %v", err)
//		return err
//	}
//
//	// Get modification timestamp from form data
//	modTimeStr := c.FormValue("modification_time")
//	if modTimeStr == "" {
//		return echo.NewHTTPError(http.StatusBadRequest, "modification_time is required")
//	}
//
//	modTime, err := strconv.ParseInt(modTimeStr, 10, 64)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusBadRequest, "invalid modification_time format")
//	}
//
//	// Get user from context
//	user, err := helpers.GetUserFromContext(c)
//	if err != nil {
//		return err
//	}
//
//	logger.HandlerTroubleReports().Info("User %s is rolling back trouble report %d to modification %d",
//		user.Name, id, modTime)
//
//	// Find the specific modification
//	modifications, err := h.DB.Modifications.ListAll(
//		services.ModificationTypeTroubleReport, id)
//	if err != nil {
//		logger.HandlerTroubleReports().Error("Failed to get modifications: %v", err)
//		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
//			"failed to retrieve modifications: "+err.Error())
//	}
//
//	var targetMod *models.Modification[interface{}]
//	for _, mod := range modifications {
//		if mod.CreatedAt.UnixMilli() == modTime {
//			targetMod = mod
//			break
//		}
//	}
//
//	if targetMod == nil {
//		return echo.NewHTTPError(http.StatusNotFound, "modification not found")
//	}
//
//	// Unmarshal the modification data
//	var modData models.TroubleReportModData
//	if err := json.Unmarshal(targetMod.Data, &modData); err != nil {
//		logger.HandlerTroubleReports().Error("Failed to unmarshal modification data: %v", err)
//		return echo.NewHTTPError(http.StatusInternalServerError,
//			"failed to parse modification data: "+err.Error())
//	}
//
//	// Get the current trouble report
//	tr, err := h.DB.TroubleReports.Get(id)
//	if err != nil {
//		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
//			"failed to retrieve trouble report: "+err.Error())
//	}
//
//	// Apply the rollback
//	tr.Title = modData.Title
//	tr.Content = modData.Content
//	tr.LinkedAttachments = modData.LinkedAttachments
//
//	// Update the trouble report
//	if err := h.DB.TroubleReports.Update(tr, user); err != nil {
//		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
//			"failed to rollback trouble report: "+err.Error())
//	}
//
//	logger.HandlerTroubleReports().Info("Successfully rolled back trouble report %d", id)
//
//	// Redirect back to modifications page
//	c.Response().Header().Set("HX-Redirect", fmt.Sprintf("/trouble-reports/modifications/%d", id))
//	return c.NoContent(http.StatusOK)
//}
