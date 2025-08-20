package constants

// IDs for HTML elements
const (
	// Dialogs
	IDDialogLogin             = "dialogLogin"
	IDDialogEditUserName      = "dialogEditUserName"
	IDToolEditDialog          = "toolEditDialog"
	IDTroubleReportEditDialog = "troubleReportEditDialog"

	// Form inputs and fields
	IDTitle                      = "title"
	IDContent                    = "content"
	IDPosition                   = "position"
	IDWidth                      = "width"
	IDHeight                     = "height"
	IDType                       = "type"
	IDCode                       = "code"
	IDAttachments                = "attachments"
	IDExistingAttachmentsRemoval = "existing-attachments-removal"

	// Lists and data containers
	IDAllToolsList        = "allToolsList"
	IDData                = "data"
	IDCookies             = "cookies"
	IDModificationsList   = "modifications-list"
	IDExistingAttachments = "existing-attachments"
	IDNewAttachments      = "new-attachments"

	// Navigation and UI elements
	IDFeedCounter        = "feedCounter"
	IDVoteForDelete      = "voteForDelete"
	IDAttachmentsSection = "attachments-section"
	IDFilePreview        = "file-preview"

	// Dynamic ID prefixes (used with fmt.Sprintf)
	IDTroubleReportPrefix = "trouble-report-%d"
)
