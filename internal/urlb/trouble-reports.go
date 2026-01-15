package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// TroubleReports constructs trouble reports page URL
func TroubleReports() templ.SafeURL {
	return BuildURL("/trouble-reports")
}

// TroubleReportsSharePDF constructs trouble reports share PDF URL
func TroubleReportsSharePDF(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/share-pdf", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsAttachment constructs trouble reports attachment URL
func TroubleReportsAttachment(attachment string) templ.SafeURL {
	params := map[string]string{}
	if attachment != "" {
		params["attachment"] = attachment
	}
	return BuildURLWithParams("/trouble-reports/attachment", params)
}

// TroubleReportsData constructs trouble reports data URL
func TroubleReportsData() templ.SafeURL {
	return BuildURL("/trouble-reports/data")
}

// TroubleReportsDelete constructs trouble reports data URL
func TroubleReportsDelete(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/delete", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsAttachmentsPreview constructs trouble reports attachments preview URL
func TroubleReportsAttachmentsPreview(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/attachments-preview", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsRollback constructs trouble reports rollback URL
func TroubleReportsRollback(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/rollback", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}
