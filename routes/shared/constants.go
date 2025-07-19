// Package shared provides shared constants and configuration for all route handlers.
//
// This package centralizes commonly used template paths, form field names, and other
// constants to reduce duplication across route handlers and ensure consistency.
// It's separated from the main routes package to avoid import cycles.
package shared

// Template file paths
const (
	LayoutTemplatePath  = "templates/layout.html"
	NavFeedTemplatePath = "templates/nav/feed.html"

	HomeTemplatePath           = "templates/home.html"
	LoginTemplatePath          = "templates/login.html"
	FeedTemplatePath           = "templates/feed.html"
	ProfileTemplatePath        = "templates/profile.html"
	TroubleReportsTemplatePath = "templates/trouble-reports.html"

	FeedDataTemplatePath                    = "templates/feed/data.html"
	FeedCounterTemplatePath                 = "templates/nav/feed-counter.html"
	ProfileCookiesTemplatePath              = "templates/profile/cookies.html"
	TroubleReportsDataTemplatePath          = "templates/trouble-reports/data.html"
	TroubleReportsModificationsTemplatePath = "templates/trouble-reports/modifications.html"
	TroubleReportsDialogTemplatePath        = "templates/trouble-reports/dialog-edit.html"
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
	IDQueryParam     = "id"
	CancelQueryParam = "cancel"
	PageQueryParam   = "page"
	LimitQueryParam  = "limit"
	SearchQueryParam = "q"
)

// Validation constants
const (
	UserNameMinLength = 1
	UserNameMaxLength = 100

	TitleMinLength   = 1
	TitleMaxLength   = 500
	ContentMinLength = 1
	ContentMaxLength = 50000

	MaxSearchQueryLength = 500
)

// Form values
const (
	TrueValue  = "true"
	FalseValue = "false"
)

// Error messages
const (
	AuthenticationRequiredMessage  = "authentication required"
	AdminPrivilegesRequiredMessage = "administrator privileges required"
	InvalidParameterMessage        = "invalid parameter"
	ValidationFailedMessage        = "validation failed"
	TemplateParseErrorMessage      = "failed to parse templates"
	TemplateExecuteErrorMessage    = "failed to render page"
)
