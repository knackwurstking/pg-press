package troublereports

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/pdf"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func GetSharePDF(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	// Get trouble report by ID
	tr, merr := db.GetTroubleReport(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	b, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return echo.NewHTTPError(500, "Fehler beim Generieren des PDFs").SetInternal(err)
	}

	return shareResponse(c, tr, b)
}

func shareResponse(c echo.Context, tr *shared.TroubleReport, buf *bytes.Buffer) *echo.HTTPError {
	if buf == nil || buf.Len() == 0 {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"PDF buffer is empty",
		)
	}

	sanitizedTitle := sanitizeFilename(tr.Title)
	if sanitizedTitle == "" {
		sanitizedTitle = "fehlerbericht"
	}

	filename := fmt.Sprintf("%d_%s_%s.pdf", tr.ID, sanitizedTitle, time.Now().Format(shared.DateFormat))

	// Set headers
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Response().Header().Set("Cache-Control", "private, max-age=0, no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("X-Frame-Options", "DENY")
	c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
	c.Response().Header().Set("Content-Description", "Trouble Report PDF")

	if err := c.Blob(http.StatusOK, "application/pdf", buf.Bytes()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

func sanitizeFilename(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}

	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "-", "_")
	filename = strings.ReplaceAll(filename, "(", "_")
	filename = strings.ReplaceAll(filename, ")", "_")
	filename = strings.ReplaceAll(filename, "[", "_")
	filename = strings.ReplaceAll(filename, "]", "_")

	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	filename = strings.Trim(filename, "_")

	if filename == "" {
		filename = "attachment"
	}

	return filename
}
