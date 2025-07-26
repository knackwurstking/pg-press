package troublereports

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/pgvis/logger"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type TemplateData struct {
	TroubleReports []*pgvis.TroubleReportWithAttachments `json:"trouble_reports"`
	User           *pgvis.User                           `json:"user"`
}

type DataHandler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func (h *DataHandler) RegisterRoutes(e *echo.Echo) {
	dataPath := h.serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	attachmentsPreviewPath := h.serverPathPrefix + "/trouble-reports/attachments-preview"
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)

	sharePdfPath := h.serverPathPrefix + "/trouble-reports/share-pdf"
	e.GET(sharePdfPath, h.handleGetSharePdf)
}

func (h *DataHandler) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.db.TroubleReportService.ListWithAttachments()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		TemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.templates,
		[]string{
			constants.TroubleReportsDataComponentTemplatePath,
		},
	)
}

func (h *DataHandler) handleDeleteData(c echo.Context) error {
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

	if err := h.db.TroubleReportService.RemoveWithAttachments(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}

type AttachmentsPreviewTemplateData struct {
	TroubleReport *pgvis.TroubleReportWithAttachments `json:"trouble_report"`
}

func (h *DataHandler) handleGetAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	tr, err := h.db.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		AttachmentsPreviewTemplateData{
			TroubleReport: tr,
		},
		h.templates,
		[]string{
			constants.TroubleReportsAttachmentsPreviewComponentTemplatePath,
		},
	)
}

func (h *DataHandler) handleGetSharePdf(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	logger.TroubleReport().Info("Generating PDF for trouble report %d", id)

	tr, err := h.db.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		logger.TroubleReport().Error("Failed to retrieve trouble report %d for PDF generation: %v", id, err)
		return utils.HandlePgvisError(c, err)
	}

	// Create PDF with UTF-8 support for German characters
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Create Unicode translator for German character support
	translator := pdf.UnicodeTranslatorFromDescriptor("")

	// Set encoding for better German character support
	pdf.SetAutoPageBreak(true, 25)

	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	// Header with title and ID
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 51, 102)
	pdf.Cell(0, 15, translator("Fehlerbericht"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 8, fmt.Sprintf("Report-ID: #%d", tr.ID))
	pdf.Ln(15)

	// Reset text color to black
	pdf.SetTextColor(0, 0, 0)

	// Title section with border
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "TITEL", "1", 1, "L", true, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 8, translator(tr.Title), "LR", "", false)
	pdf.CellFormat(0, 0, "", "T", 1, "", false, 0, "")
	pdf.Ln(8)

	// Content section with border
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "INHALT", "1", 1, "L", true, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, translator(tr.Content), "LR", "", false)
	pdf.CellFormat(0, 0, "", "T", 1, "", false, 0, "")
	pdf.Ln(8)

	// Metadata section
	if len(tr.Mods) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(240, 248, 255)
		pdf.CellFormat(0, 10, "METADATEN", "1", 1, "L", true, 0, "")

		// Find creation and last modification times
		var earliestTime, latestTime int64 = tr.Mods[0].Time, tr.Mods[0].Time
		var creator, lastModifier *pgvis.User

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
		pdf.MultiCell(0, 6, translator(createdText), "LR", "", false)

		if latestTime != earliestTime {
			lastModifiedAt := time.Unix(0, latestTime*int64(time.Millisecond))
			modifiedText := fmt.Sprintf("Zuletzt geändert: %s", lastModifiedAt.Format("02.01.2006 15:04:05"))
			if lastModifier != nil {
				modifiedText += fmt.Sprintf(" von %s", lastModifier.UserName)
			}
			pdf.MultiCell(0, 6, translator(modifiedText), "LR", "", false)
		}

		pdf.Cell(0, 6, translator(fmt.Sprintf("Anzahl Änderungen: %d", len(tr.Mods))))
		pdf.Ln(5)
		pdf.CellFormat(0, 0, "", "T", 1, "", false, 0, "")
		pdf.Ln(8)
	}

	// Attachments section
	if len(tr.LoadedAttachments) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(240, 248, 255)
		pdf.CellFormat(0, 10, translator(fmt.Sprintf("ANHÄNGE (%d)", len(tr.LoadedAttachments))), "1", 1, "L", true, 0, "")

		pdf.SetFont("Arial", "", 11)

		// Calculate column width for side-by-side layout
		pageWidth, _ := pdf.GetPageSize()
		leftMargin, _, rightMargin, _ := pdf.GetMargins()
		availableWidth := pageWidth - leftMargin - rightMargin
		columnWidth := availableWidth / 2
		leftColumnX := leftMargin
		rightColumnX := leftMargin + columnWidth

		// Track the current row's Y position and maximum height
		var currentRowY float64
		var maxRowHeight float64
		_, currentRowY = pdf.GetXY()

		for i, attachment := range tr.LoadedAttachments {
			// Determine which column (0 = left, 1 = right)
			column := i % 2
			isLeftColumn := column == 0

			// Set X position for current column
			var columnX float64
			if isLeftColumn {
				columnX = leftColumnX
				// For left column, start a new row if not the first attachment
				if i > 0 {
					currentRowY += maxRowHeight + 8 // Add spacing between rows
					maxRowHeight = 0                // Reset for new row
				}
			} else {
				columnX = rightColumnX
			}

			// Set position for this attachment
			pdf.SetXY(columnX, currentRowY)
			attachmentStartY := currentRowY

			pdf.Cell(columnWidth-5, 6, translator(fmt.Sprintf("• Anhang %d", i+1)))
			pdf.SetXY(columnX, attachmentStartY+6)
			pdf.Cell(columnWidth-5, 6, fmt.Sprintf("Typ: %s", attachment.GetMimeType()))

			// Handle image attachments - embed actual images
			if attachment.IsImage() {
				logger.TroubleReport().Info("Processing image attachment %d (MIME: %s) for PDF", attachment.GetID(), attachment.GetMimeType())
				pdf.SetXY(columnX, attachmentStartY+12)
				pdf.Cell(columnWidth-5, 6, "Kategorie: Bild")

				// Fetch actual image data
				attachmentData, err := h.db.Attachments.Get(attachment.GetID())
				if err != nil {
					logger.TroubleReport().Error("Failed to fetch attachment %d for PDF: %v", attachment.GetID(), err)
					pdf.SetXY(columnX, attachmentStartY+18)
					pdf.Cell(columnWidth-5, 6, "[Bild konnte nicht geladen werden]")
					// Update row height for this attachment
					currentAttachmentHeight := 24.0
					if currentAttachmentHeight > maxRowHeight {
						maxRowHeight = currentAttachmentHeight
					}
				} else {
					logger.TroubleReport().Debug("Successfully fetched attachment %d data (%d bytes)", attachment.GetID(), len(attachmentData.Data))
					// Determine image type from MIME type
					var imageType string
					switch attachment.GetMimeType() {
					case "image/jpeg", "image/jpg":
						imageType = "JPG"
					case "image/png":
						imageType = "PNG"
					case "image/gif":
						imageType = "GIF"
					default:
						// Try to use JPG as fallback for other formats
						imageType = "JPG"
					}

					// Create temporary image name for PDF registration
					imageName := fmt.Sprintf("attachment_%d", attachment.GetID())
					logger.TroubleReport().Debug("Attempting to register image %s as type %s", imageName, imageType)

					// Try to register image with error handling
					err := func() error {
						defer func() {
							if r := recover(); r != nil {
								logger.TroubleReport().Error("Panic while registering image %d: %v", attachment.GetID(), r)
							}
						}()
						pdf.RegisterImageReader(imageName, imageType, bytes.NewReader(attachmentData.Data))
						return nil
					}()

					if err != nil {
						logger.TroubleReport().Error("Failed to register image %d: %v", attachment.GetID(), err)
						pdf.SetXY(columnX, attachmentStartY+18)
						pdf.Cell(columnWidth-5, 6, "[Bild-Format fehlerhaft]")
						// Update row height for this attachment
						currentAttachmentHeight := 24.0
						if currentAttachmentHeight > maxRowHeight {
							maxRowHeight = currentAttachmentHeight
						}
						continue
					}

					logger.TroubleReport().Debug("Successfully registered image %s in PDF", imageName)

					// Calculate image dimensions for column layout
					maxWidth := columnWidth - 5 // Maximum width per column (reduced margin for bigger images)
					maxHeight := 80.0           // Maximum height in mm (increased for bigger images)

					// Position for image
					imageY := attachmentStartY + 18

					// Get image dimensions and calculate scaled size with error handling
					imageInfo := func() *gofpdf.ImageInfoType {
						defer func() {
							if r := recover(); r != nil {
								logger.TroubleReport().Error("Panic while getting image info for %d: %v", attachment.GetID(), r)
							}
						}()
						return pdf.GetImageInfo(imageName)
					}()

					if imageInfo != nil {
						// Original dimensions in points (convert to mm)
						origWidthPt, origHeightPt := imageInfo.Extent()
						origWidthMM := origWidthPt * 25.4 / 72
						origHeightMM := origHeightPt * 25.4 / 72
						logger.TroubleReport().Debug("Image %d original dimensions: %.1fx%.1fmm", attachment.GetID(), origWidthMM, origHeightMM)

						// Calculate scale to fit within max dimensions
						scaleW := maxWidth / origWidthMM
						scaleH := maxHeight / origHeightMM
						scale := scaleW
						if scaleH < scaleW {
							scale = scaleH
						}

						// Apply scale
						imgWidth := origWidthMM * scale
						imgHeight := origHeightMM * scale
						logger.TroubleReport().Debug("Image %d scaled dimensions: %.1fx%.1fmm (scale: %.2f)", attachment.GetID(), imgWidth, imgHeight, scale)

						// Ensure minimum readable size
						if imgWidth < 20 {
							imgWidth = 20
							imgHeight = origHeightMM * (20 / origWidthMM)
						}

						// Check if image fits on current page, add new page if needed
						_, pageHeight := pdf.GetPageSize()
						_, _, _, bottomMargin := pdf.GetMargins()
						if imageY+imgHeight+10 > pageHeight-bottomMargin {
							pdf.AddPage()
							imageY = pdf.GetY()
						}

						// Center the image within the column
						imageX := columnX + (columnWidth-imgWidth)/2

						// Add the image with proper error handling
						err := func() error {
							defer func() {
								if r := recover(); r != nil {
									logger.TroubleReport().Error("Panic while adding image %d to PDF: %v", attachment.GetID(), r)
								}
							}()
							pdf.ImageOptions(imageName, imageX, imageY, imgWidth, imgHeight, false, gofpdf.ImageOptions{
								ImageType: imageType,
							}, 0, "")
							return nil
						}()

						if err != nil {
							logger.TroubleReport().Error("Failed to insert image %d into PDF: %v", attachment.GetID(), err)
							pdf.SetXY(columnX, imageY)
							pdf.Cell(columnWidth-5, 6, translator("[Bild konnte nicht eingefügt werden]"))
							// Update row height for this attachment
							currentAttachmentHeight := 30.0
							if currentAttachmentHeight > maxRowHeight {
								maxRowHeight = currentAttachmentHeight
							}
						} else {
							logger.TroubleReport().Info("Successfully embedded image attachment %d in PDF", attachment.GetID())

							// Add image caption below image
							pdf.SetFont("Arial", "I", 8)
							pdf.SetXY(columnX, imageY+imgHeight+2)
							pdf.Cell(columnWidth-5, 4, translator(fmt.Sprintf("Anhang %d (%s)", i+1, attachment.GetMimeType())))
							pdf.SetFont("Arial", "", 11) // Reset font

							// Update row height to include image + caption
							currentAttachmentHeight := imgHeight + 20 // 18 (header) + imgHeight + 2 (caption spacing)
							if currentAttachmentHeight > maxRowHeight {
								maxRowHeight = currentAttachmentHeight
							}
						}
					} else {
						logger.TroubleReport().Warn("Could not get image info for attachment %d, format may not be supported", attachment.GetID())
						pdf.SetXY(columnX, imageY)
						pdf.Cell(columnWidth-5, 6, translator("[Bild-Format nicht unterstützt]"))
						// Update row height for this attachment
						currentAttachmentHeight := 30.0
						if currentAttachmentHeight > maxRowHeight {
							maxRowHeight = currentAttachmentHeight
						}
					}
				}
			} else if attachment.IsDocument() {
				pdf.SetXY(columnX, attachmentStartY+12)
				pdf.Cell(columnWidth-5, 6, "Kategorie: Dokument")
				// Update row height for non-image attachments
				currentAttachmentHeight := 18.0
				if currentAttachmentHeight > maxRowHeight {
					maxRowHeight = currentAttachmentHeight
				}
			} else if attachment.IsArchive() {
				pdf.SetXY(columnX, attachmentStartY+12)
				pdf.Cell(columnWidth-5, 6, "Kategorie: Archiv")
				// Update row height for non-image attachments
				currentAttachmentHeight := 18.0
				if currentAttachmentHeight > maxRowHeight {
					maxRowHeight = currentAttachmentHeight
				}
			} else {
				pdf.SetXY(columnX, attachmentStartY+12)
				pdf.Cell(columnWidth-5, 6, "Kategorie: Andere")
				// Update row height for non-image attachments
				currentAttachmentHeight := 18.0
				if currentAttachmentHeight > maxRowHeight {
					maxRowHeight = currentAttachmentHeight
				}
			}
		}

		// Position cursor after all attachments
		pdf.SetXY(leftMargin, currentRowY+maxRowHeight+8)
		pdf.CellFormat(0, 0, "", "T", 1, "", false, 0, "")
		pdf.Ln(5)
	}

	// Create buffer and write PDF
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		logger.TroubleReport().Error("Failed to generate PDF for trouble report %d: %v", tr.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Fehler beim Erstellen der PDF")
	}

	logger.TroubleReport().Info("Successfully generated PDF for trouble report %d (size: %d bytes)", tr.ID, buf.Len())

	// Set headers for PDF download and Web Share API compatibility
	filename := fmt.Sprintf("fehlerbericht_%d_%s.pdf", tr.ID, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Methods", "GET")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}
