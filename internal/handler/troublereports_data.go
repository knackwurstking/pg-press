package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/logger"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

func (h *TroubleReports) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.DB.TroubleReportService.ListWithAttachments()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		TroubleReportsDataTemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.Templates,
		[]string{
			constants.TroubleReportsDataComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handleDeleteData(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
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

	logger.TroubleReport().Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if err := h.DB.TroubleReportService.RemoveWithAttachments(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		AttachmentsPreviewTemplateData{
			TroubleReport: tr,
		},
		h.Templates,
		[]string{
			constants.TroubleReportsAttachmentsPreviewComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handleGetSharePdf(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	logger.TroubleReport().Info("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		logger.TroubleReport().Error(
			"Failed to retrieve trouble report %d for PDF generation: %v", id, err)
		return utils.HandlePgvisError(c, err)
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	translator := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 51, 102)
	pdf.Cell(0, 15, translator("Fehlerbericht"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 8, fmt.Sprintf("Report-ID: #%d", tr.ID))
	pdf.Ln(15)

	pdf.SetTextColor(0, 0, 0)

	// Title section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "TITEL", "1", 1, "L", true, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 8, translator(tr.Title), "", "", false)
	pdf.Ln(8)

	// Content section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "INHALT", "1", 1, "L", true, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, translator(tr.Content), "", "", false)
	pdf.Ln(8)

	// Metadata
	if len(tr.Mods) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(240, 248, 255)
		pdf.CellFormat(0, 10, "METADATEN", "1", 1, "L", true, 0, "")
		pdf.Ln(5)

		var earliestTime, latestTime int64 = tr.Mods[0].Time, tr.Mods[0].Time
		var creator, lastModifier *database.User

		for _, mod := range tr.Mods {
			if mod.Time < earliestTime {
				earliestTime = mod.Time
				creator = mod.User
			}
			if mod.Time > latestTime {
				latestTime = mod.Time
				lastModifier = mod.User
			}
		}

		pdf.SetFont("Arial", "", 11)
		createdAt := time.Unix(0, earliestTime*int64(time.Millisecond))
		createdText := fmt.Sprintf("Erstellt am: %s", createdAt.Format("02.01.2006 15:04:05"))
		if creator != nil {
			createdText += fmt.Sprintf(" von %s", creator.UserName)
		}
		pdf.MultiCell(0, 6, translator(createdText), "", "", false)

		if latestTime != earliestTime {
			lastModifiedAt := time.Unix(0, latestTime*int64(time.Millisecond))
			modifiedText := fmt.Sprintf("Zuletzt geändert: %s",
				lastModifiedAt.Format("02.01.2006 15:04:05"))
			if lastModifier != nil {
				modifiedText += fmt.Sprintf(" von %s", lastModifier.UserName)
			}
			pdf.MultiCell(0, 6, translator(modifiedText), "", "", false)
		}

		pdf.Cell(0, 6, translator(fmt.Sprintf("Anzahl Änderungen: %d", len(tr.Mods))))
		pdf.Ln(13)
	}

	// Attachments
	if len(tr.LoadedAttachments) > 0 {
		// Collect only image attachments
		var images []*database.Attachment

		for _, attachment := range tr.LoadedAttachments {
			if attachment.IsImage() {
				images = append(images, attachment)
			}
		}

		// Only proceed if there are images
		if len(images) > 0 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 14)
			pdf.SetFillColor(240, 248, 255)
			pdf.CellFormat(0, 10,
				translator(fmt.Sprintf("BILDER (%d)", len(images))),
				"1", 1, "L", true, 0, "")

			// Display images 2 per row
			pageWidth, _ := pdf.GetPageSize()
			leftMargin, _, rightMargin, _ := pdf.GetMargins()
			usableWidth := pageWidth - leftMargin - rightMargin
			imageWidth := (usableWidth - 10) / 2 // 10mm spacing between images

			pdf.Ln(10)

			// Process images in pairs (2 per row)
			var currentY float64
			_, currentY = pdf.GetXY()

			for i := 0; i < len(images); i += 2 {
				// First, determine the heights of both images in this row
				var leftHeight, rightHeight float64
				var leftTmpFile, rightTmpFile string
				var leftImageType, rightImageType string

				// Process left image to get height
				leftImage := images[i]
				tmpFile1, err := os.CreateTemp("", fmt.Sprintf("attachment_%s_*.jpg", leftImage.ID))
				if err == nil {
					_, err = tmpFile1.Write(leftImage.Data)
					tmpFile1.Close()
					leftTmpFile = tmpFile1.Name()

					if err == nil {
						// Determine image type from mime type
						switch leftImage.MimeType {
						case "image/jpeg", "image/jpg":
							leftImageType = "JPG"
						case "image/png":
							leftImageType = "PNG"
						case "image/gif":
							leftImageType = "GIF"
						default:
							leftImageType = "JPG" // Default fallback
						}

						// Get left image height by registering the image
						info := pdf.RegisterImage(leftTmpFile, leftImageType)
						if info != nil {
							leftHeight = (imageWidth * info.Height()) / info.Width()
						}
					}
				}

				// Process right image to get height (if it exists)
				if i+1 < len(images) {
					rightImage := images[i+1]
					tmpFile2, err := os.CreateTemp("", fmt.Sprintf("attachment_%s_*.jpg", rightImage.ID))
					if err == nil {
						_, err = tmpFile2.Write(rightImage.Data)
						tmpFile2.Close()
						rightTmpFile = tmpFile2.Name()

						if err == nil {
							// Determine image type from mime type
							switch rightImage.MimeType {
							case "image/jpeg", "image/jpg":
								rightImageType = "JPG"
							case "image/png":
								rightImageType = "PNG"
							case "image/gif":
								rightImageType = "GIF"
							default:
								rightImageType = "JPG" // Default fallback
							}

							// Get right image height by registering the image
							info := pdf.RegisterImage(rightTmpFile, rightImageType)
							if info != nil {
								rightHeight = (imageWidth * info.Height()) / info.Width()
							}
						}
					}
				}

				// Determine the actual row height (max of both images)
				actualRowHeight := leftHeight
				if rightHeight > actualRowHeight {
					actualRowHeight = rightHeight
				}
				if actualRowHeight == 0 {
					actualRowHeight = 60.0 // Fallback height if images couldn't be processed
				}

				// Calculate positions for this row
				captionY := currentY
				imageY := captionY + 6
				rightX := leftMargin + imageWidth + 10

				// Check if we need a new page before adding the images
				if imageY+actualRowHeight+25 > 270 {
					pdf.AddPage()
					_, currentY = pdf.GetXY()
					captionY = currentY
					imageY = captionY + 6
				}

				// Add captions for both images
				pdf.SetFont("Arial", "", 9)

				// Left image caption
				pdf.SetXY(leftMargin, captionY)
				pdf.CellFormat(imageWidth, 4,
					translator(fmt.Sprintf("Anhang %d", i+1)),
					"0", 0, "C", false, 0, "")

				// Right image caption (if exists)
				if i+1 < len(images) {
					pdf.SetXY(rightX, captionY)
					pdf.CellFormat(imageWidth, 4,
						translator(fmt.Sprintf("Anhang %d", i+2)),
						"0", 0, "C", false, 0, "")
				}

				// Add left image
				if leftTmpFile != "" {
					pdf.Image(leftTmpFile, leftMargin, imageY, imageWidth, 0, false, leftImageType, 0, "")
					os.Remove(leftTmpFile)
				}

				// Add right image (if it exists)
				if rightTmpFile != "" {
					pdf.Image(rightTmpFile, rightX, imageY, imageWidth, 0, false, rightImageType, 0, "")
					os.Remove(rightTmpFile)
				}

				// Move to next row using actual row height
				currentY = imageY + actualRowHeight + 15
				pdf.SetXY(leftMargin, currentY)
			}
		}
	}

	// Create buffer and write PDF
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		logger.TroubleReport().Error("Failed to generate PDF for trouble report %d: %v", tr.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Fehler beim Erstellen der PDF")
	}

	logger.TroubleReport().Info(
		"Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, buf.Len())

	filename := fmt.Sprintf("fehlerbericht_%d_%s.pdf",
		tr.ID, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Length",
		fmt.Sprintf("%d", buf.Len()))

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}

func (h *TroubleReports) handleGetModifications(c echo.Context, tr *database.TroubleReport) *echo.HTTPError {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	if tr == nil {
		var err error
		tr, err = h.DB.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	loadedAttachments, err := h.DB.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	mods := slices.Clone(tr.Mods)
	slices.Reverse(mods)

	data := &ModificationsTemplateData{
		User:              user,
		TroubleReport:     tr,
		LoadedAttachments: loadedAttachments,
		Mods:              mods,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{
			constants.TroubleReportsModificationsComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handlePostModifications(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Move modification to the top
	newMods := []*database.Modified[database.TroubleReportMod]{}
	var mod *database.Modified[database.TroubleReportMod]
	for _, m := range tr.Mods {
		if m.Time == timeQuery {
			if mod != nil {
				logger.TroubleReport().Warn(
					"Multiple modifications with the same time, mod: %+v, m: %+v", mod, m)
				newMods = append(newMods, m)
			} else {
				mod = m
			}
		} else {
			newMods = append(newMods, m)
		}
	}

	if mod == nil {
		return utils.HandlePgvisError(c, errors.New("modification not found"))
	}

	mod.Time = time.Now().UnixMilli()

	// Update mods with new order
	tr.Mods = append(newMods, mod)

	// Update trouble reports data
	tr.Title = mod.Data.Title
	tr.Content = mod.Data.Content
	tr.LinkedAttachments = mod.Data.LinkedAttachments

	// Update database
	if err = h.DB.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, database.WrapError(err, "failed to update trouble report"))
	}

	return h.handleGetModifications(c, tr)
}

func (h *TroubleReports) handleGetModificationAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find the specific modification
	var targetMod *database.Modified[database.TroubleReportMod]
	for _, mod := range tr.Mods {
		if mod.Time == timeQuery {
			targetMod = mod
			break
		}
	}

	if targetMod == nil {
		return utils.HandlePgvisError(c, errors.New("modification not found"))
	}

	// Load attachments for this modification
	var attachments []*database.Attachment
	if len(targetMod.Data.LinkedAttachments) > 0 {
		attachments, err = h.DB.Attachments.GetByIDs(targetMod.Data.LinkedAttachments)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	data := &ModificationAttachmentsTemplateData{
		TroubleReport: tr,
		Modification:  targetMod,
		Attachments:   attachments,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{
			constants.TroubleReportsModificationAttachmentsPreviewComponentTemplatePath,
		},
	)
}
