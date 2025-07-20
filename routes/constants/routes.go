// Package constants provides shared constants for the routes package.
package constants

import "time"

// Cookie configuration
const (
	CookieName               = "pgvis-api-key"
	CookieExpirationDuration = time.Hour * 24 * 31 * 6
)

// Template file paths
const (
	LayoutTemplatePath  = "templates/layouts/main.html"
	NavFeedTemplatePath = "templates/components/nav/feed.html"

	HomeTemplatePath           = "templates/pages/home.html"
	LoginTemplatePath          = "templates/pages/login.html"
	FeedTemplatePath           = "templates/pages/feed.html"
	ProfileTemplatePath        = "templates/pages/profile.html"
	TroubleReportsTemplatePath = "templates/pages/trouble-reports.html"

	FeedDataTemplatePath                    = "templates/components/feed/data.html"
	FeedCounterTemplatePath                 = "templates/components/nav/feed-counter.html"
	ProfileCookiesTemplatePath              = "templates/components/profile/cookies.html"
	TroubleReportsDataTemplatePath          = "templates/components/trouble-reports/data.html"
	TroubleReportsModificationsTemplatePath = "templates/components/trouble-reports/modifications.html"
	TroubleReportsDialogTemplatePath        = "templates/components/trouble-reports/dialog-edit.html"
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
	//PageQueryParam   = "page"
	//LimitQueryParam  = "limit"
	//SearchQueryParam = "q"
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
	TrueValue = "true"
	//FalseValue = "false"
)

// Error messages
const (
	RedirectFailedMessage = "failed to redirect"
)
