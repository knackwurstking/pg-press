package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// TroubleReportsPage constructs trouble reports page URL
func TroubleReportsPage() templ.SafeURL {
	return BuildURL("/trouble-reports")
}

// TroubleReportsSharePDF constructs trouble reports share PDF URL
func TroubleReportsSharePDF(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/share-pdf", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsAttachment constructs trouble reports attachment URL
func TroubleReportsAttachment(trID, aID shared.EntityID, modificationTime int64) templ.SafeURL {
	params := map[string]string{}
	if trID != 0 {
		params["id"] = fmt.Sprintf("%d", trID)
	}
	if aID != 0 {
		params["attachment_id"] = fmt.Sprintf("%d", aID)
	}
	if modificationTime != 0 {
		params["modification_time"] = fmt.Sprintf("%d", modificationTime)
	}
	return BuildURLWithParams("/trouble-reports/attachment", params)
}

// TroubleReportsModifications constructs trouble reports modifications URL
func TroubleReportsModifications(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams(fmt.Sprintf("/trouble-reports/modifications/%d", trID), map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsData constructs trouble reports data URL
func TroubleReportsData(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/data", map[string]string{
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
