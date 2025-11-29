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
			if v != "" {
				values.Add(k, v)
			}
		}
		if len(values) > 0 {
			u = fmt.Sprintf("%s?%s", u, values.Encode())
		}
	}

	return templ.SafeURL(u)
}

func UrlLogin(apiKey string, invalid *bool) (url struct {
	Page templ.SafeURL
}) {
	params := map[string]string{}
	if apiKey != "" {
		params["api-key"] = apiKey
	}
	if invalid != nil {
		params["invalid"] = fmt.Sprintf("%t", *invalid)
	}
	url.Page = BuildURL("/login", params)
	return url
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
	List templ.SafeURL
}) {
	url.Page = BuildURL("/feed", nil)
	url.List = BuildURL("/feed/list", nil)
	return url
}

func UrlHelp() (url struct {
	MarkdownPage templ.SafeURL
}) {
	url.MarkdownPage = BuildURL("/help/markdown", nil)
	return url
}

func UrlEditor(_type, id, returnURL string) (url struct {
	Page templ.SafeURL
	Save templ.SafeURL
}) {
	url.Page = BuildURL("/editor", map[string]string{
		"type":      _type,
		"id":        id,
		"returnURL": returnURL,
	})

	url.Save = BuildURL("/editor/save", nil)

	return url
}

func UrlProfile(value string) (url struct {
	Page    templ.SafeURL
	Cookies templ.SafeURL
}) {
	url.Page = BuildURL("/profile", nil)
	url.Cookies = BuildURL("/profile", map[string]string{
		"value": value,
	})
	return url
}

func UrlNotes(noteID string) (url struct {
	Page   templ.SafeURL
	Delete templ.SafeURL
	Grid   templ.SafeURL
}) {
	url.Page = BuildURL("/notes", nil)
	url.Delete = BuildURL("/notes/delete", map[string]string{
		"id": noteID,
	})
	url.Grid = BuildURL("/notes/grid", nil)
	return url
}

func UrlMetalSheets(metalSheetID string) (url struct {
	Delete templ.SafeURL
}) {
	url.Delete = BuildURL("/metal-sheets/delete", map[string]string{
		"id": metalSheetID,
	})
	return url
}

func UrlUmbau(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/umbau/%d", press), nil)
	return url
}

func UrlTroubleReports(troubleReportID, attachmentID, modificationTime string) (url struct {
	Page               templ.SafeURL
	SharePDF           templ.SafeURL
	Attachment         templ.SafeURL
	Modifications      templ.SafeURL
	Data               templ.SafeURL
	AttachmentsPreview templ.SafeURL
	Rollback           templ.SafeURL
}) {
	params := map[string]string{
		"id":                troubleReportID,
		"attachment_id":     attachmentID,
		"modification_time": modificationTime,
	}

	url.Page = BuildURL("/trouble-reports", nil)
	url.SharePDF = BuildURL("/trouble-reports/share-pdf", params)
	url.Attachment = BuildURL("/trouble-reports/attachment", params)
	url.Modifications = BuildURL("/trouble-reports/modifications/"+troubleReportID, params)
	url.Data = BuildURL("/trouble-reports/data", params)
	url.AttachmentsPreview = BuildURL("/trouble-reports/attachments-preview", params)
	url.Rollback = BuildURL("/trouble-reports/rollback", params)

	return url
}

// TODO: Fix this across the whole project and also check the handler for params and routes
func UrlTools() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("/tools", nil)
	return url
}

// TODO: Fix this across the whole project and also check the handler for params and routes
func UrlTool(tool models.ToolID) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/tool/%d", tool), nil)
	return url
}

// TODO: Fix this across the whole project and also check the handler for params and routes
func UrlPress(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press/%d", press), nil)
	return url
}

// TODO: Fix this across the whole project and also check the handler for params and routes
func UrlPressRegeneration(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press-regeneration/%d", press), nil)
	return url
}
