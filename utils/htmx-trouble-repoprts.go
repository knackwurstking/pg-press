package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/env"
)

func HXGetTroubleReportsData() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/data",
		env.ServerPathPrefix,
	))
}

func HXDeleteTroubleReportsData(troubleReportID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/data?id=%d",
		env.ServerPathPrefix, troubleReportID,
	))
}

func HXGetTroubleReportsAttachmentsPreview(troubleReportID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/attachments-preview?id=%d",
		env.ServerPathPrefix, troubleReportID,
	))
}

func HXPostTroubleReportsRollback(troubleReportID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/rollback?id=%d",
		env.ServerPathPrefix, troubleReportID,
	))
}
