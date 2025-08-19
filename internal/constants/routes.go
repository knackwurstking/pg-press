// Package constants provides shared constants for the routes package.
package constants

import "time"

// Cookie configuration
const (
	CookieName               = "pgpress-api-key"
	CookieExpirationDuration = time.Hour * 24 * 31 * 6 // 6 months
)

// Form field names
const (
	APIKeyFormField            = "api-key"
	UserNameFormField          = "user-name"
	TitleFormField             = "title"
	ContentFormField           = "content"
	AttachmentsFormField       = "attachments"
	ExistingAttachmentsRemoval = "existing_attachments_removal"
)

// Query parameter names
const (
	QueryParamID           = "id"
	QueryParamCancel       = "cancel"
	QueryParamTime         = "time"
	QueryParamPress        = "press"
	QueryParamAttachmentID = "attachment_id"
)

// Validation constants
const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
)

// Error messages
const (
	RedirectFailedMessage = "failed to redirect"
)
