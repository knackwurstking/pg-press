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
	cookiesSorted := []*pgvis.Cookie{}

outer:
	for _, c := range p.Cookies {
		for i, sc := range cookiesSorted {
			if c.LastLogin > sc.LastLogin {
				cookiesSorted = slices.Insert(cookiesSorted, i, c)
				continue outer
			}
		}

		cookiesSorted = append(cookiesSorted, c)
	}

	return cookiesSorted
}
