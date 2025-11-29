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

func UrlProfile(cookieValue string) (url struct {
	Page    templ.SafeURL
	Cookies templ.SafeURL
}) {
	url.Page = BuildURL("/profile", nil)
	url.Cookies = BuildURL("/profile", map[string]string{
		"value": cookieValue,
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

func UrlTools(id string) (url struct {
	Page                  templ.SafeURL
	Delete                templ.SafeURL
	MarkDead              templ.SafeURL
	SectionPress          templ.SafeURL
	SectionTools          templ.SafeURL
	AdminOverlappingTools templ.SafeURL
}) {
	params := map[string]string{
		"id": id,
	}

	url.Page = BuildURL("/tools", nil)
	url.Delete = BuildURL("/tools/delete", params)
	url.MarkDead = BuildURL("/tools/mark-dead", params)
	url.SectionPress = BuildURL("/tools/section-press", nil)
	url.SectionTools = BuildURL("/tools/section-tools", nil)
	url.AdminOverlappingTools = BuildURL("/tools/admin-overlapping-tools", nil)

	return url
}

func UrlTool(toolID, toolRegenerationID, cycleID string) (url struct {
	Page               templ.SafeURL
	DeleteRegeneration templ.SafeURL
	StatusEdit         templ.SafeURL
	StatusDisplay      templ.SafeURL
	Status             templ.SafeURL
	Notes              templ.SafeURL
	MetalSheets        templ.SafeURL
	Cycles             templ.SafeURL
	TotalCycles        templ.SafeURL
	CycleDelete        templ.SafeURL
	Bind               templ.SafeURL
	UnBind             templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/tool/%s", toolID), nil)

	url.DeleteRegeneration = BuildURL(fmt.Sprintf("/tool/%s/delete-regeneration", toolID), map[string]string{
		"id": toolRegenerationID,
	})

	url.StatusEdit = BuildURL(fmt.Sprintf("/tool/%s/status-edit", toolID), nil)
	url.StatusDisplay = BuildURL(fmt.Sprintf("/tool/%s/status-display", toolID), nil)
	url.Status = BuildURL(fmt.Sprintf("/tool/%s/status", toolID), nil)
	url.Notes = BuildURL(fmt.Sprintf("/tool/%s/notes", toolID), nil)
	url.MetalSheets = BuildURL(fmt.Sprintf("/tool/%s/metal-sheets", toolID), nil)
	url.Cycles = BuildURL(fmt.Sprintf("/tool/%s/cycles", toolID), nil)
	url.TotalCycles = BuildURL(fmt.Sprintf("/tool/%s/total-cycles", toolID), nil)

	url.CycleDelete = BuildURL("/tool/cycle/delete", map[string]string{
		"id": cycleID,
	})

	url.Bind = BuildURL(fmt.Sprintf("/tool/%s/bind", toolID), nil)
	url.UnBind = BuildURL(fmt.Sprintf("/tool/%s/unbind", toolID), nil)

	return url
}

func UrlPress(pressNumber string) (url struct {
	Page            templ.SafeURL
	ActiveTools     templ.SafeURL
	MetalSheets     templ.SafeURL
	Cycles          templ.SafeURL
	Notes           templ.SafeURL
	CycleSummaryPDF templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press/%s", pressNumber), nil)
	url.ActiveTools = BuildURL(fmt.Sprintf("/press/%s/active-tools", pressNumber), nil)
	url.MetalSheets = BuildURL(fmt.Sprintf("/press/%s/metal-sheets", pressNumber), nil)
	url.Cycles = BuildURL(fmt.Sprintf("/press/%s/cycles", pressNumber), nil)
	url.Notes = BuildURL(fmt.Sprintf("/press/%s/notes", pressNumber), nil)
	url.CycleSummaryPDF = BuildURL(fmt.Sprintf("/press/%s/cycle-summary-pdf", pressNumber), nil)

	return url
}

func UrlPressRegeneration(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press-regeneration/%d", press), nil)
	return url
}

func UrlDialogs() (url struct {
	EditCycle            func(cycleID, toolChangeMode string) templ.SafeURL
	EditTool             func(toolID string) templ.SafeURL
	EditMetalSheet       func(metalSheetID, toolID string) templ.SafeURL
	EditNote             func(noteID, linkToTables string) templ.SafeURL
	EditToolRegeneration func(toolRegenerationID string) templ.SafeURL
}) {
	url.EditCycle = urlEditCycleDialog
	url.EditTool = urlEditToolDialog
	url.EditMetalSheet = urlEditMetalSheetDialog
	url.EditNote = urlEditNoteDialog
	url.EditToolRegeneration = urlEditToolRegenerationDialog

	return url
}

func urlEditCycleDialog(cycleID, toolChangeMode string) templ.SafeURL {
	params := map[string]string{
		"id":               cycleID,
		"tool_change_mode": toolChangeMode,
	}

	return BuildURL("/dialog/edit-cycle", params)
}

func urlEditToolDialog(toolID string) templ.SafeURL {
	params := map[string]string{
		"id": toolID,
	}

	return BuildURL("/dialog/edit-tool", params)
}

func urlEditMetalSheetDialog(metalSheetID, toolID string) templ.SafeURL {
	params := map[string]string{
		"id":      metalSheetID,
		"tool_id": toolID,
	}

	return BuildURL("/dialog/edit-metal-sheet", params)
}

func urlEditNoteDialog(noteID, linkToTables string) templ.SafeURL {
	params := map[string]string{
		"id":             noteID,
		"link_to_tables": linkToTables,
	}

	return BuildURL("/dialog/edit-note", params)
}

func urlEditToolRegenerationDialog(toolRegenerationID string) templ.SafeURL {
	params := map[string]string{
		"id": toolRegenerationID,
	}

	return BuildURL("/dialog/edit-tool-regeneration", params)
}
