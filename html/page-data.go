package html

import (
	"slices"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
)

type PageData struct {
	ErrorMessages []string
}

type LoginPageData struct {
	PageData

	ApiKey        string
	InvalidApiKey bool
}

type ProfilePageData struct {
	PageData

	User    *pgvis.User
	Cookies []*pgvis.Cookie
}

func (p ProfilePageData) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}
