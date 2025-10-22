package utils

import (
	"strconv"

	"github.com/a-h/templ"
)

func HXGetTroubleReportsData() templ.SafeURL {
	return buildURL("/htmx/trouble-reports/data", nil)
}

func HXDeleteTroubleReportsData(troubleReportID int64) templ.SafeURL {
	return buildURL("/htmx/trouble-reports/data", map[string]string{
		"id": strconv.FormatInt(troubleReportID, 10),
	})
}

func HXGetTroubleReportsAttachmentsPreview(troubleReportID int64) templ.SafeURL {
	return buildURL("/htmx/trouble-reports/attachments-preview", map[string]string{
		"id": strconv.FormatInt(troubleReportID, 10),
	})
}

func HXPostTroubleReportsRollback(troubleReportID int64) templ.SafeURL {
	return buildURL("/htmx/trouble-reports/rollback", map[string]string{
		"id": strconv.FormatInt(troubleReportID, 10),
	})
}
