package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/models"
)

func HXGetTroubleReportsData() templ.SafeURL {
	return buildURL("/htmx/trouble-reports/data", nil)
}

func HXDeleteTroubleReportsData(troubleReportID models.TroubleReportID) templ.SafeURL {
	return buildURL("/htmx/trouble-reports/data", map[string]string{
		"id": fmt.Sprintf("%d", troubleReportID),
	})
}

func HXGetTroubleReportsAttachmentsPreview(troubleReportID models.TroubleReportID) templ.SafeURL {
	return buildURL("/htmx/trouble-reports/attachments-preview", map[string]string{
		"id": fmt.Sprintf("%d", troubleReportID),
	})
}

func HXPostTroubleReportsRollback(troubleReportID models.TroubleReportID) templ.SafeURL {
	return buildURL("/htmx/trouble-reports/rollback", map[string]string{
		"id": fmt.Sprintf("%d", troubleReportID),
	})
}
