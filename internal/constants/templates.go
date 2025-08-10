// Package constants provides shared template constants for the routes package.
package constants

// New template paths (target structure)
const (
	// Layout templates
	MainLayoutTemplatePath = "templates/layouts/main.html"

	ProfilePageTemplatePath        = "templates/pages/profile.html"
	TroubleReportsPageTemplatePath = "templates/pages/trouble-reports.html"

	// Component templates
	NavFeedComponentTemplatePath = "templates/components/nav-feed.html"
)

const (
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
