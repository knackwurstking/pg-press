// Package constants provides shared constants for the application.
package constants

// HTML element IDs for dialogs
const (
	// IDDialogLogin is the ID for the login dialog element
	IDDialogLogin = "dialogLogin"

	// IDDialogEditUserName is the ID for the edit user name dialog element
	IDDialogEditUserName = "dialogEditUserName"

	// IDToolEditDialog is the ID for the tool edit dialog element
	IDToolEditDialog = "toolEditDialog"

	// IDTroubleReportEditDialog is the ID for the trouble report edit dialog element
	IDTroubleReportEditDialog = "troubleReportEditDialog"
)

// HTML element IDs for form inputs and fields
const (
	// IDTitle is the ID for the title input element
	IDTitle = "title"

	// IDContent is the ID for the content input element
	IDContent = "content"

	// IDPosition is the ID for the position input element
	IDPosition = "position"

	// IDWidth is the ID for the width input element
	IDWidth = "width"

	// IDHeight is the ID for the height input element
	IDHeight = "height"

	// IDType is the ID for the type input element
	IDType = "type"

	// IDCode is the ID for the code input element
	IDCode = "code"

	// IDAttachments is the ID for the attachments input element
	IDAttachments = "attachments"

	// IDExistingAttachmentsRemoval is the ID for the existing attachments removal element
	IDExistingAttachmentsRemoval = "existing-attachments-removal"
)

// HTML element IDs for lists and data containers
const (
	// IDAllToolsList is the ID for the all tools list element
	IDAllToolsList = "allToolsList"

	// IDData is the ID for the data container element
	IDData = "data"

	// IDCookies is the ID for the cookies container element
	IDCookies = "cookies"

	// IDModificationsList is the ID for the modifications list element
	IDModificationsList = "modifications-list"

	// IDExistingAttachments is the ID for the existing attachments container element
	IDExistingAttachments = "existing-attachments"

	// IDNewAttachments is the ID for the new attachments container element
	IDNewAttachments = "new-attachments"
)

// HTML element IDs for navigation and UI elements
const (
	// IDFeedCounter is the ID for the feed counter element
	IDFeedCounter = "feedCounter"

	// IDVoteForDelete is the ID for the vote for delete element
	IDVoteForDelete = "voteForDelete"

	// IDAttachmentsSection is the ID for the attachments section element
	IDAttachmentsSection = "attachments-section"

	// IDFilePreview is the ID for the file preview element
	IDFilePreview = "file-preview"
)

// Dynamic ID prefixes for generating HTML element IDs
const (
	// IDTroubleReportPrefix is the prefix for trouble report IDs (used with fmt.Sprintf)
	IDTroubleReportPrefix = "trouble-report-%d"
)
