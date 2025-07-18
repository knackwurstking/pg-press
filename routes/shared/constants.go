// Package shared provides shared constants and configuration for all route handlers.
//
// This package centralizes commonly used template paths, form field names, and other
// constants to reduce duplication across route handlers and ensure consistency.
// It's separated from the main routes package to avoid import cycles.
package shared

// Common template file paths used across multiple route handlers
const (
	// Core layout templates
	LayoutTemplatePath  = "templates/layout.html"
	NavFeedTemplatePath = "templates/nav/feed.html"

	// Main page templates
	HomeTemplatePath           = "templates/home.html"
	LoginTemplatePath          = "templates/login.html"
	FeedTemplatePath           = "templates/feed.html"
	ProfileTemplatePath        = "templates/profile.html"
	TroubleReportsTemplatePath = "templates/trouble-reports.html"

	// Data/component templates
	FeedDataTemplatePath             = "templates/feed/data.html"
	FeedCounterTemplatePath          = "templates/nav/feed-counter.html"
	ProfileCookiesTemplatePath       = "templates/profile/cookies.html"
	TroubleReportsDataTemplatePath   = "templates/trouble-reports/data.html"
	TroubleReportsDialogTemplatePath = "templates/trouble-reports/dialog-edit.html"
)

// Common form field names used across different forms
const (
	// Authentication and user management
	APIKeyFormField   = "api-key"
	UserNameFormField = "user-name"

	// Trouble reports
	TitleFormField   = "title"
	ContentFormField = "content"
)

// Common query parameter names
const (
	// General parameters
	IDQueryParam     = "id"
	CancelQueryParam = "cancel"

	// Pagination parameters
	PageQueryParam  = "page"
	LimitQueryParam = "limit"

	// Search parameters
	SearchQueryParam = "q"
)

// Validation constants for form fields
const (
	// Username validation
	UserNameMinLength = 1
	UserNameMaxLength = 100

	// Trouble report validation
	TitleMinLength   = 1
	TitleMaxLength   = 500
	ContentMinLength = 1
	ContentMaxLength = 50000

	// General validation
	MaxSearchQueryLength = 500
)

// Special form values
const (
	TrueValue  = "true"
	FalseValue = "false"
)

// HTTP response constants
const (
	// Common error messages
	AuthenticationRequiredMessage  = "authentication required"
	AdminPrivilegesRequiredMessage = "administrator privileges required"
	InvalidParameterMessage        = "invalid parameter"
	ValidationFailedMessage        = "validation failed"

	// Template error messages
	TemplateParseErrorMessage   = "failed to parse templates"
	TemplateExecuteErrorMessage = "failed to render page"
)
