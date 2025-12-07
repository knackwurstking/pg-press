package utils

import (
	"fmt"
	"net/url"
	"strings"

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

// UrlLogin constructs login URL with optional API key and invalid flag
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

// UrlNav constructs navigation URLs
func UrlNav() (url struct {
	FeedCounter templ.SafeURL
}) {
	url.FeedCounter = BuildURL("/nav/feed-counter", nil)
	return url
}

// UrlHome constructs home URL
func UrlHome() (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL("", nil)
	return url
}

// UrlFeed constructs feed URLs
func UrlFeed() (url struct {
	Page templ.SafeURL
	List templ.SafeURL
}) {
	url.Page = BuildURL("/feed", nil)
	url.List = BuildURL("/feed/list", nil)
	return url
}

// UrlHelp constructs help URLs
func UrlHelp() (url struct {
	MarkdownPage templ.SafeURL
}) {
	url.MarkdownPage = BuildURL("/help/markdown", nil)
	return url
}

// UrlEditor constructs editor URLs
func UrlEditor(_type models.EditorType, id string, returnURL templ.SafeURL) (url struct {
	Page templ.SafeURL
	Save templ.SafeURL
}) {
	a, _ := strings.CutPrefix(string(returnURL), env.ServerPathPrefix)
	url.Page = BuildURL("/editor", map[string]string{
		"type":       string(_type),
		"id":         id,
		"return_url": string(a),
	})

	url.Save = BuildURL("/editor/save", nil)

	return url
}

// UrlProfile constructs profile URLs
func UrlProfile(cookieValue string) (url struct {
	Page    templ.SafeURL
	Cookies templ.SafeURL
}) {
	url.Page = BuildURL("/profile", nil)
	url.Cookies = BuildURL("/profile/cookies", map[string]string{
		"value": cookieValue,
	})
	return url
}

// UrlNotes constructs notes URLs
func UrlNotes(noteID models.NoteID) (url struct {
	Page   templ.SafeURL
	Delete templ.SafeURL
	Grid   templ.SafeURL
}) {
	url.Page = BuildURL("/notes", nil)
	url.Delete = BuildURL("/notes/delete", map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	})
	url.Grid = BuildURL("/notes/grid", nil)
	return url
}

// UrlMetalSheets constructs metal sheets URLs
func UrlMetalSheets(metalSheetID models.MetalSheetID) (url struct {
	Delete templ.SafeURL
}) {
	url.Delete = BuildURL("/metal-sheets/delete", map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	})
	return url
}

// UrlUmbau constructs umbau URLs
func UrlUmbau(press models.PressNumber) (url struct {
	Page templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/umbau/%d", press), nil)
	return url
}

// UrlTroubleReports constructs trouble reports URLs
func UrlTroubleReports(trID models.TroubleReportID, aID models.AttachmentID, modificationTime int64) (url struct {
	Page               templ.SafeURL
	SharePDF           templ.SafeURL
	Attachment         templ.SafeURL
	Modifications      templ.SafeURL
	Data               templ.SafeURL
	AttachmentsPreview templ.SafeURL
	Rollback           templ.SafeURL
}) {
	params := map[string]string{}
	if trID != 0 {
		params["id"] = fmt.Sprintf("%d", trID)
	}
	if aID != 0 {
		params["attachment_id"] = fmt.Sprintf("%d", aID)
	}
	if modificationTime != 0 {
		params["modification_time"] = fmt.Sprintf("%d", modificationTime)
	}

	url.Page = BuildURL("/trouble-reports", nil)
	url.SharePDF = BuildURL("/trouble-reports/share-pdf", params)
	url.Attachment = BuildURL("/trouble-reports/attachment", params)
	url.Modifications = BuildURL(fmt.Sprintf("/trouble-reports/modifications/%d", trID), params)
	url.Data = BuildURL("/trouble-reports/data", params)
	url.AttachmentsPreview = BuildURL("/trouble-reports/attachments-preview", params)
	url.Rollback = BuildURL("/trouble-reports/rollback", params)

	return url
}

// UrlTools constructs tools URLs
func UrlTools(id models.ToolID) (url struct {
	Page                  templ.SafeURL
	Delete                templ.SafeURL
	MarkDead              templ.SafeURL
	SectionPress          templ.SafeURL
	SectionTools          templ.SafeURL
	AdminOverlappingTools templ.SafeURL
}) {
	params := map[string]string{}
	if id != 0 {
		params["id"] = fmt.Sprintf("%d", id)
	}

	url.Page = BuildURL("/tools", nil)
	url.Delete = BuildURL("/tools/delete", params)
	url.MarkDead = BuildURL("/tools/mark-dead", params)
	url.SectionPress = BuildURL("/tools/section/press", nil)
	url.SectionTools = BuildURL("/tools/section/tools", nil)
	url.AdminOverlappingTools = BuildURL("/tools/admin/overlapping-tools", nil)

	return url
}

// UrlTool constructs tool URLs
func UrlTool(toolID models.ToolID, toolRegenerationID models.ToolRegenerationID, cycleID models.CycleID) (url struct {
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
	url.Page = BuildURL(fmt.Sprintf("/tool/%d", toolID), nil)

	{
		params := map[string]string{}
		if toolRegenerationID != 0 {
			params["id"] = fmt.Sprintf("%d", toolRegenerationID)
		}
		url.DeleteRegeneration = BuildURL(
			fmt.Sprintf("/tool/%d/delete-regeneration", toolID), params)
	}

	url.StatusEdit = BuildURL(fmt.Sprintf("/tool/%d/status-edit", toolID), nil)
	url.StatusDisplay = BuildURL(fmt.Sprintf("/tool/%d/status-display", toolID), nil)
	url.Status = BuildURL(fmt.Sprintf("/tool/%d/status", toolID), nil)
	url.Notes = BuildURL(fmt.Sprintf("/tool/%d/notes", toolID), nil)
	url.MetalSheets = BuildURL(fmt.Sprintf("/tool/%d/metal-sheets", toolID), nil)
	url.Cycles = BuildURL(fmt.Sprintf("/tool/%d/cycles", toolID), nil)
	url.TotalCycles = BuildURL(fmt.Sprintf("/tool/%d/total-cycles", toolID), nil)

	{
		params := map[string]string{}
		if cycleID != 0 {
			params["id"] = fmt.Sprintf("%d", cycleID)
		}
		url.CycleDelete = BuildURL("/tool/cycle/delete", params)
	}

	url.Bind = BuildURL(fmt.Sprintf("/tool/%d/bind", toolID), nil)
	url.UnBind = BuildURL(fmt.Sprintf("/tool/%d/unbind", toolID), nil)

	return url
}

// UrlPress constructs press URLs
func UrlPress(pressNumber models.PressNumber) (url struct {
	Page               templ.SafeURL
	ActiveTools        templ.SafeURL
	MetalSheets        templ.SafeURL
	Cycles             templ.SafeURL
	Notes              templ.SafeURL
	PressRegenerations templ.SafeURL
	CycleSummaryPDF    templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press/%d", pressNumber), nil)
	url.ActiveTools = BuildURL(fmt.Sprintf("/press/%d/active-tools", pressNumber), nil)
	url.MetalSheets = BuildURL(fmt.Sprintf("/press/%d/metal-sheets", pressNumber), nil)
	url.Cycles = BuildURL(fmt.Sprintf("/press/%d/cycles", pressNumber), nil)
	url.Notes = BuildURL(fmt.Sprintf("/press/%d/notes", pressNumber), nil)
	url.PressRegenerations = BuildURL(fmt.Sprintf("/press/%d/press-regenerations", pressNumber), nil)
	url.CycleSummaryPDF = BuildURL(fmt.Sprintf("/press/%d/cycle-summary-pdf", pressNumber), nil)

	return url
}

// UrlPressRegeneration constructs press regeneration URLs
func UrlPressRegeneration(press models.PressNumber, rid models.PressRegenerationID) (url struct {
	Page   templ.SafeURL
	Delete templ.SafeURL
}) {
	params := map[string]string{
		"id": fmt.Sprintf("%d", rid),
	}

	url.Page = BuildURL(fmt.Sprintf("/press-regeneration/%d", press), nil)
	url.Delete = BuildURL(fmt.Sprintf("/press-regeneration/%d/delete", press), params)

	return url
}

// UrlDialogs constructs dialog URLs
func UrlDialogs() (url struct {
	EditCycle             func(cycleID models.CycleID, toolID models.ToolID, toolChangeMode *bool) templ.SafeURL
	EditTool              func(toolID models.ToolID) templ.SafeURL
	EditMetalSheet        func(metalSheetID models.MetalSheetID, toolID models.ToolID) templ.SafeURL
	EditNote              func(noteID models.NoteID, linkToTables string) templ.SafeURL
	EditToolRegeneration  func(toolRegenerationID models.ToolRegenerationID) templ.SafeURL
	EditPressRegeneration func(pressRegenerationID models.PressRegenerationID) templ.SafeURL
}) {
	url.EditCycle = urlEditCycleDialog
	url.EditTool = urlEditToolDialog
	url.EditMetalSheet = urlEditMetalSheetDialog
	url.EditNote = urlEditNoteDialog
	url.EditToolRegeneration = urlEditToolRegenerationDialog
	url.EditPressRegeneration = urlEditPressRegenerationDialog

	return url
}

// urlEditCycleDialog constructs edit cycle dialog URL
func urlEditCycleDialog(cycleID models.CycleID, toolID models.ToolID, toolChangeMode *bool) templ.SafeURL {
	params := map[string]string{}
	if cycleID != 0 {
		params["id"] = fmt.Sprintf("%d", cycleID)
	}
	if toolID != 0 {
		params["tool_id"] = fmt.Sprintf("%d", toolID)
	}
	if toolChangeMode != nil {
		params["tool_change_mode"] = fmt.Sprintf("%t", *toolChangeMode)
	}

	return BuildURL("/dialog/edit-cycle", params)
}

// urlEditToolDialog constructs edit tool dialog URL
func urlEditToolDialog(toolID models.ToolID) templ.SafeURL {
	params := map[string]string{}
	if toolID != 0 {
		params["id"] = fmt.Sprintf("%d", toolID)
	}

	return BuildURL("/dialog/edit-tool", params)
}

// urlEditMetalSheetDialog constructs edit metal sheet dialog URL
func urlEditMetalSheetDialog(metalSheetID models.MetalSheetID, toolID models.ToolID) templ.SafeURL {
	params := map[string]string{}
	if metalSheetID != 0 {
		params["id"] = fmt.Sprintf("%d", metalSheetID)
	}
	if toolID != 0 {
		params["tool_id"] = fmt.Sprintf("%d", toolID)
	}

	return BuildURL("/dialog/edit-metal-sheet", params)
}

// urlEditNoteDialog constructs edit note dialog URL
func urlEditNoteDialog(noteID models.NoteID, linkToTables string) templ.SafeURL {
	params := map[string]string{
		"link_to_tables": linkToTables,
	}
	if noteID != 0 {
		params["id"] = fmt.Sprintf("%d", noteID)
	}

	return BuildURL("/dialog/edit-note", params)
}

// urlEditToolRegenerationDialog constructs edit tool regeneration dialog URL
func urlEditToolRegenerationDialog(toolRegenerationID models.ToolRegenerationID) templ.SafeURL {
	params := map[string]string{}
	if toolRegenerationID != 0 {
		params["id"] = fmt.Sprintf("%d", toolRegenerationID)
	}

	return BuildURL("/dialog/edit-tool-regeneration", params)
}

// urlEditPressRegenerationDialog constructs edit press regeneration dialog URL
func urlEditPressRegenerationDialog(pressRegenerationID models.PressRegenerationID) templ.SafeURL {
	params := map[string]string{}
	if pressRegenerationID != 0 {
		params["id"] = fmt.Sprintf("%d", pressRegenerationID)
	}

	return BuildURL("/dialog/edit-press-regeneration", params)
}
