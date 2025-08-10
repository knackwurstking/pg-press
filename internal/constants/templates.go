// Package constants provides shared template constants for the routes package.
package constants

// New template paths (target structure)
const (
	// Layout templates
	MainLayoutTemplatePath = "templates/layouts/main.html"

	// Page templates
	FeedPageTemplatePath           = "templates/pages/feed.html"
	ProfilePageTemplatePath        = "templates/pages/profile.html"
	TroubleReportsPageTemplatePath = "templates/pages/trouble-reports.html"

	// Component templates
	NavFeedComponentTemplatePath = "templates/components/nav-feed.html"
)

const (
	// Feed component templates
	HTMXFeedDataTemplatePath = "templates/htmx/feed/data.html"

	// Profile component templates
	HTMXProfileCookiesTemplatePath = "templates/htmx/profile/cookies.html"

	// Trouble Reports component templates
	HTMXTroubleReportsDataTemplatePath               = "templates/htmx/trouble-reports/data.html"
	HTMXTroubleReportsModificationsTemplatePath      = "templates/htmx/trouble-reports/modifications.html"
	HTMXTroubleReportsDialogEditTemplatePath         = "templates/htmx/trouble-reports/dialog-edit.html"
	HTMXTroubleReportsAttachmentsPreviewTemplatePath = "templates/htmx/trouble-reports/attachments-preview-container.html"
)

// Template sets for common page combinations
var (
	FeedPageTemplates = []string{
		MainLayoutTemplatePath,
		NavFeedComponentTemplatePath,
		FeedPageTemplatePath,
	}

	ProfilePageTemplates = []string{
		MainLayoutTemplatePath,
		NavFeedComponentTemplatePath,
		ProfilePageTemplatePath,
	}

	TroubleReportsPageTemplates = []string{
		MainLayoutTemplatePath,
		NavFeedComponentTemplatePath,
		TroubleReportsPageTemplatePath,
	}
)
