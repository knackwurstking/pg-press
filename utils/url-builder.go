package utils

import (
	"fmt"
	"net/url"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
)

// BuildURL constructs a URL with the given path and query parameters
func BuildURL(path string, params map[string]string) templ.SafeURL {
	u := fmt.Sprintf("%s%s", env.ServerPathPrefix, path)

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		u = fmt.Sprintf("%s?%s", u, values.Encode())
	}

	return templ.SafeURL(u)
}

func UrlNav() (url struct {
	FeedCounter templ.SafeURL
}) {
	url.FeedCounter = BuildURL("/nav/feed-counter", nil)
	return url
}

func UrlHome() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("", nil)
	return url
}

func UrlFeed() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/feed", nil)
	return url
}

func UrlHelp() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/help", nil)
	return url
}

func UrlEditor() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/editor", nil)
	return url
}

func UrlProfile() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/profile", nil)
	return url
}

func UrlNotes() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/notes", nil)
	return url
}

func UrlMetalSheets() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/metal-sheets", nil)
	return url
}

func UrlUmbau(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/umbau/%d", press), nil)
	return url
}

func UrlTroubleReports() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/trouble-reports", nil)
	return url
}

func UrlTools() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/tools", nil)
	return url
}

func UrlTool(tool models.ToolID) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/tool/%d", tool), nil)
	return url
}

func UrlPress(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press/%d", press), nil)
	return url
}

func UrlPressRegeneration(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press-regeneration/%d", press), nil)
	return url
}
