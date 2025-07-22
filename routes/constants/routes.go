// Package constants provides shared constants for the routes package.
package constants

import "time"

// Cookie configuration
const (
	CookieName               = "pgvis-api-key"
	CookieExpirationDuration = time.Hour * 24 * 31 * 6
)

// Form field names
const (
	APIKeyFormField   = "api-key"
	UserNameFormField = "user-name"
	TitleFormField    = "title"
	ContentFormField  = "content"
)

// Query parameter names
const (
	QueryParamID     = "id"
	QueryParamCancel = "cancel"
	QueryParamTime   = "time"
	//PageQueryParam   = "page"
	//LimitQueryParam  = "limit"
	//SearchQueryParam = "q"
)

// Validation constants
const (
	UserNameMinLength = 1
	UserNameMaxLength = 100

	//TitleMinLength   = 1
	//TitleMaxLength   = 500
	//ContentMinLength = 1
	//ContentMaxLength = 50000

	//MaxSearchQueryLength = 500
)

// Form values
const (
	TrueValue = "true"
	//FalseValue = "false"
)

// Error messages
const (
	RedirectFailedMessage = "failed to redirect"
)
