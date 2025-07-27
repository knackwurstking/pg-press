// Package constants provides shared constants for the routes package.
package constants

import "time"

// Cookie configuration
const (
	CookieName               = "pgvis-api-key"
	CookieExpirationDuration = time.Hour * 24 * 31 * 6 // 6 months
)

// Form field names
const (
	APIKeyFormField          = "api-key"
	UserNameFormField        = "user-name"
	TitleFormField           = "title"
	ContentFormField         = "content"
	AttachmentsFormField     = "attachments"
	AttachmentOrderField     = "attachment_order"
	ExistingAttachmentPrefix = "existing_attachment_"
)

// Query parameter names
const (
	QueryParamID     = "id"
	QueryParamCancel = "cancel"
	QueryParamTime   = "time"
)

// Validation constants
const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
)

// Form values
const (
	TrueValue = "true"
)

// Error messages
const (
	RedirectFailedMessage = "failed to redirect"
)
