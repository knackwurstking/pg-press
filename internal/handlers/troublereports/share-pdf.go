package troublereports

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/pdf"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

// TODO: Continue here
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

	return nil
}
