// Package constants provides shared template constants for the routes package.
package constants

// New template paths (target structure)
const (
	// Layout templates
	MainLayoutTemplatePath = "templates/layouts/main.html"

	// Page templates
	HomePageTemplatePath           = "templates/pages/home.html"
	LoginPageTemplatePath          = "templates/pages/login.html"
	FeedPageTemplatePath           = "templates/pages/feed.html"
	ProfilePageTemplatePath        = "templates/pages/profile.html"
	TroubleReportsPageTemplatePath = "templates/pages/trouble-reports.html"

	// Component templates
	NavFeedComponentTemplatePath                          = "templates/components/nav/feed.html"
	FeedCounterComponentTemplatePath                      = "templates/components/nav/feed-counter.html"
	FeedDataComponentTemplatePath                         = "templates/components/feed/data.html"
	ProfileCookiesComponentTemplatePath                   = "templates/components/profile/cookies.html"
	TroubleReportsDataComponentTemplatePath               = "templates/components/trouble-reports/data.html"
	TroubleReportsModificationsComponentTemplatePath      = "templates/components/trouble-reports/modifications.html"
	TroubleReportsDialogEditComponentTemplatePath         = "templates/components/trouble-reports/dialog-edit.html"
	TroubleReportsAttachmentsPreviewComponentTemplatePath = "templates/components/trouble-reports/attachments-preview.html"
	AttachmentsSectionComponentTemplatePath               = "templates/components/attachments/section.html"
)

// Template sets for common page combinations
var (
	HomePageTemplates = []string{
		MainLayoutTemplatePath,
		NavFeedComponentTemplatePath,
		HomePageTemplatePath,
	}

	LoginPageTemplates = []string{
		MainLayoutTemplatePath,
		NavFeedComponentTemplatePath,
		LoginPageTemplatePath,
	}

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
