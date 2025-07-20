// Package constants provides shared template constants for the routes package.
package constants

// Legacy template paths (for backward compatibility during migration)
const (
	// Current template paths that need to be updated
	LegacyLayoutTemplatePath  = "templates/layout.html"
	LegacyNavFeedTemplatePath = "templates/nav/feed.html"

	LegacyHomeTemplatePath           = "templates/home.html"
	LegacyLoginTemplatePath          = "templates/login.html"
	LegacyFeedTemplatePath           = "templates/feed.html"
	LegacyProfileTemplatePath        = "templates/profile.html"
	LegacyTroubleReportsTemplatePath = "templates/trouble-reports.html"

	LegacyFeedDataTemplatePath                    = "templates/feed/data.html"
	LegacyFeedCounterTemplatePath                 = "templates/nav/feed-counter.html"
	LegacyProfileCookiesTemplatePath              = "templates/profile/cookies.html"
	LegacyTroubleReportsDataTemplatePath          = "templates/trouble-reports/data.html"
	LegacyTroubleReportsModificationsTemplatePath = "templates/trouble-reports/modifications.html"
	LegacyTroubleReportsDialogTemplatePath        = "templates/trouble-reports/dialog-edit.html"
)

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
	NavFeedComponentTemplatePath                     = "templates/components/nav/feed.html"
	FeedCounterComponentTemplatePath                 = "templates/components/nav/feed-counter.html"
	FeedDataComponentTemplatePath                    = "templates/components/feed/data.html"
	ProfileCookiesComponentTemplatePath              = "templates/components/profile/cookies.html"
	TroubleReportsDataComponentTemplatePath          = "templates/components/trouble-reports/data.html"
	TroubleReportsModificationsComponentTemplatePath = "templates/components/trouble-reports/modifications.html"
	TroubleReportsDialogComponentTemplatePath        = "templates/components/trouble-reports/dialog-edit.html"
)

// Template sets for common page combinations
var (
	HomePageTemplates = []string{
		LegacyLayoutTemplatePath,
		LegacyHomeTemplatePath,
		LegacyNavFeedTemplatePath,
	}

	LoginPageTemplates = []string{
		LegacyLayoutTemplatePath,
		LegacyLoginTemplatePath,
	}

	FeedPageTemplates = []string{
		LegacyLayoutTemplatePath,
		LegacyFeedTemplatePath,
	}

	ProfilePageTemplates = []string{
		LegacyLayoutTemplatePath,
		LegacyProfileTemplatePath,
	}

	TroubleReportsPageTemplates = []string{
		LegacyLayoutTemplatePath,
		LegacyTroubleReportsTemplatePath,
	}
)
