// Package constants provides shared constants for the application.
package constants

// HTML element IDs for dialogs
const (
	IDDialogLogin             = "dialogLogin"
	IDDialogEditUserName      = "dialogEditUserName"
	IDToolEditDialog          = "toolEditDialog"
	IDToolCycleEditDialog     = "toolCycleEditDialog"
	IDTroubleReportEditDialog = "troubleReportEditDialog"
)

// HTML element IDs for form inputs and fields
const (
	IDTitle                      = "title"
	IDContent                    = "content"
	IDPosition                   = "position"
	IDPressSelection             = "pressSelection"
	IDWidth                      = "width"
	IDHeight                     = "height"
	IDType                       = "type"
	IDCode                       = "code"
	IDAttachments                = "attachments"
	IDExistingAttachmentsRemoval = "existing-attachments-removal"
)

// HTML element IDs for lists and data containers
const (
	IDAllToolsList        = "allToolsList"
	IDData                = "data"
	IDCookies             = "cookies"
	IDModificationsList   = "modifications-list"
	IDExistingAttachments = "existing-attachments"
	IDNewAttachments      = "new-attachments"
)

// HTML element IDs for navigation and UI elements
const (
	IDFeedCounter        = "feedCounter"
	IDVoteForDelete      = "voteForDelete"
	IDAttachmentsSection = "attachments-section"
	IDFilePreview        = "file-preview"
)

// Dynamic ID prefixes for generating HTML element IDs
const (
	IDTroubleReportPrefix = "trouble-report-%d"
)
