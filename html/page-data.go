package html

import (
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
)

type PageData struct {
	ErrorMessages []string
}

func (PageData) TemplatePatterns(patterns ...string) []string {
	return append(
		patterns,
		"svg/triangle-alert.html",
		"svg/pencil.html",
		"svg/plus.html",
	)
}

type LoginPageData struct {
	PageData

	ApiKey        string
	InvalidApiKey bool
}

type ProfilePageData struct {
	PageData

	User *pgvis.User
}
