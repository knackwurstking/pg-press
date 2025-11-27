package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/models"
)

func HXGetTroubleReportsData() templ.SafeURL {
	return BuildURL("/htmx/trouble-reports/data", nil)
}

func HXDeleteTroubleReportsData(troubleReportID models.TroubleReportID) templ.SafeURL {
	return BuildURL("/htmx/trouble-reports/data", map[string]string{
		"id": fmt.Sprintf("%d", troubleReportID),
	})
}

func HXGetTroubleReportsAttachmentsPreview(troubleReportID models.TroubleReportID) templ.SafeURL {
	return BuildURL("/htmx/trouble-reports/attachments-preview", map[string]string{
		"id": fmt.Sprintf("%d", troubleReportID),
	})
}

func HXPostTroubleReportsRollback(troubleReportID models.TroubleReportID) templ.SafeURL {
	return BuildURL("/htmx/trouble-reports/rollback", map[string]string{
		"id": fmt.Sprintf("%d", troubleReportID),
	})
}
