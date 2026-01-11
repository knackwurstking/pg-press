package urlb

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// URL Builders
// -----------------------------------------------------------------------------

// BuildURL constructs a URL with the given path and query parameters
func BuildURL(path string) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("%s%s", env.ServerPathPrefix, path))
}

// BuildURLWithParams constructs a URL with the given path and query parameters
func BuildURLWithParams(path string, params map[string]string) templ.SafeURL {
	values := url.Values{}
	for k, v := range params {
		if v == "" {
			continue
		}
		values.Add(k, v)
	}
	if len(values) > 0 {
		return BuildURL(fmt.Sprintf("%s?%s", path, values.Encode()))
	}
	return BuildURL(path)
}

// -----------------------------------------------------------------------------
// Auth URLs
// -----------------------------------------------------------------------------

// UrlLogin constructs login URL with optional API key and invalid flag
func UrlLogin(apiKey string, invalid *bool) templ.SafeURL {
	params := map[string]string{}
	if apiKey != "" {
		params["api-key"] = apiKey
	}
	if invalid != nil {
		params["invalid"] = fmt.Sprintf("%t", *invalid)
	}
	return BuildURLWithParams("/login", params)
}

// -----------------------------------------------------------------------------
// Home URLs
// -----------------------------------------------------------------------------

// UrlHome constructs home URL
func UrlHome() templ.SafeURL {
	return BuildURL("/")
}

// -----------------------------------------------------------------------------
// Feed URLs
// -----------------------------------------------------------------------------

// UrlFeed constructs feed page URL
func UrlFeed() templ.SafeURL {
	return BuildURL("/feed")
}

// UrlFeedList constructs feed list URL
func UrlFeedList() templ.SafeURL {
	return BuildURL("/feed/list")
}

// -----------------------------------------------------------------------------
// Help URLs
// -----------------------------------------------------------------------------

// UrlHelpMarkdown constructs help markdown URL
func UrlHelpMarkdown() templ.SafeURL {
	return BuildURL("/help/markdown")
}

// -----------------------------------------------------------------------------
// Editor URLs
// -----------------------------------------------------------------------------

// UrlEditor constructs editor page URL
func UrlEditor(_type shared.EditorType, id string, returnURL templ.SafeURL) templ.SafeURL {
	a, _ := strings.CutPrefix(string(returnURL), env.ServerPathPrefix)
	return BuildURLWithParams("/editor", map[string]string{
		"type":       string(_type),
		"id":         id,
		"return_url": string(a),
	})
}

// UrlEditorSave constructs editor save URL
func EditorSave() templ.SafeURL {
	return BuildURL("/editor/save")
}

// -----------------------------------------------------------------------------
// Profile URLs
// -----------------------------------------------------------------------------

// UrlProfile constructs profile page URL
func UrlProfile() templ.SafeURL {
	return BuildURL("/profile")
}

// UrlProfileCookies constructs profile cookies URL
func UrlProfileCookies(cookieValue string) templ.SafeURL {
	return BuildURLWithParams("/profile/cookies", map[string]string{
		"value": cookieValue,
	})
}

// -----------------------------------------------------------------------------
// Notes URLs
// -----------------------------------------------------------------------------

// UrlNotes constructs notes page URL
func UrlNotes() templ.SafeURL {
	return BuildURL("/notes")
}

// UrlNotesDelete constructs notes delete URL
func UrlNotesDelete(noteID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/notes/delete", map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	})
}

// UrlNotesGrid constructs notes grid URL
func UrlNotesGrid() templ.SafeURL {
	return BuildURL("/notes/grid")
}

// -----------------------------------------------------------------------------
// Metal Sheets URLs
// -----------------------------------------------------------------------------

// UrlMetalSheetDelete constructs metal sheet delete URL
func UrlMetalSheetDelete(metalSheetID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/metal-sheets/delete", map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	})
}

// -----------------------------------------------------------------------------
// Umbau URLs
// -----------------------------------------------------------------------------

// UrlUmbau constructs umbau URLs
func UrlUmbau(press shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/umbau/%d", press))
}

// -----------------------------------------------------------------------------
// Trouble Reports URLs
// -----------------------------------------------------------------------------

// UrlTroubleReports constructs trouble reports URLs
func UrlTroubleReports(trID shared.EntityID, aID shared.EntityID, modificationTime int64) (url struct {
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

	url.Page = BuildURL("/trouble-reports")
	url.SharePDF = BuildURLWithParams("/trouble-reports/share-pdf", params)
	url.Attachment = BuildURLWithParams("/trouble-reports/attachment", params)
	url.Modifications = BuildURLWithParams(fmt.Sprintf("/trouble-reports/modifications/%d", trID), params)
	url.Data = BuildURLWithParams("/trouble-reports/data", params)
	url.AttachmentsPreview = BuildURLWithParams("/trouble-reports/attachments-preview", params)
	url.Rollback = BuildURLWithParams("/trouble-reports/rollback", params)

	return url
}

// -----------------------------------------------------------------------------
// Tools URLs
// -----------------------------------------------------------------------------

// UrlTools constructs tools URLs
func UrlTools(toolID shared.EntityID) (url struct {
	Page                  templ.SafeURL
	Delete                templ.SafeURL
	MarkDead              templ.SafeURL
	SectionPress          templ.SafeURL
	SectionTools          templ.SafeURL
	AdminOverlappingTools templ.SafeURL
}) {
	params := map[string]string{}
	if toolID != 0 {
		params["id"] = fmt.Sprintf("%d", toolID)
	}

	url.Page = BuildURL("/tools")
	url.Delete = BuildURLWithParams("/tools/delete", params)
	url.MarkDead = BuildURLWithParams("/tools/mark-dead", params)
	url.SectionPress = BuildURL("/tools/section/press")
	url.SectionTools = BuildURL("/tools/section/tools")
	url.AdminOverlappingTools = BuildURL("/tools/admin/overlapping-tools")

	return url
}

// -----------------------------------------------------------------------------
// Tool URLs
// -----------------------------------------------------------------------------

// UrlTool constructs tool URLs
func UrlTool(toolID, toolRegenerationID, cycleID shared.EntityID) (url struct {
	Page                templ.SafeURL
	DeleteRegeneration  templ.SafeURL
	RegenerationEdit    templ.SafeURL
	RegenerationDisplay templ.SafeURL
	Regeneration        templ.SafeURL
	Notes               templ.SafeURL
	MetalSheets         templ.SafeURL
	Cycles              templ.SafeURL
	TotalCycles         templ.SafeURL
	CycleDelete         templ.SafeURL
	Bind                templ.SafeURL
	UnBind              templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/tool/%d", toolID))

	{
		params := map[string]string{}
		if toolRegenerationID != 0 {
			params["id"] = fmt.Sprintf("%d", toolRegenerationID)
		}
		url.DeleteRegeneration = BuildURLWithParams(fmt.Sprintf("/tool/%d/delete-regeneration", toolID), params)
	}

	url.RegenerationEdit = BuildURL(fmt.Sprintf("/tool/%d/regeneration-edit", toolID))
	url.RegenerationDisplay = BuildURL(fmt.Sprintf("/tool/%d/regeneration-display", toolID))
	url.Regeneration = BuildURL(fmt.Sprintf("/tool/%d/regeneration", toolID))
	url.Notes = BuildURL(fmt.Sprintf("/tool/%d/notes", toolID))
	url.MetalSheets = BuildURL(fmt.Sprintf("/tool/%d/metal-sheets", toolID))
	url.Cycles = BuildURL(fmt.Sprintf("/tool/%d/cycles", toolID))
	url.TotalCycles = BuildURL(fmt.Sprintf("/tool/%d/total-cycles", toolID))

	{
		params := map[string]string{}
		if cycleID != 0 {
			params["id"] = fmt.Sprintf("%d", cycleID)
		}
		url.CycleDelete = BuildURLWithParams("/tool/cycle/delete", params)
	}

	url.Bind = BuildURL(fmt.Sprintf("/tool/%d/bind", toolID))
	url.UnBind = BuildURL(fmt.Sprintf("/tool/%d/unbind", toolID))

	return url
}

// -----------------------------------------------------------------------------
// Press URLs
// -----------------------------------------------------------------------------

// UrlPress constructs press URLs
func UrlPress(pressNumber shared.PressNumber) (url struct {
	Page               templ.SafeURL
	ActiveTools        templ.SafeURL
	MetalSheets        templ.SafeURL
	Cycles             templ.SafeURL
	Notes              templ.SafeURL
	PressRegenerations templ.SafeURL
	CycleSummaryPDF    templ.SafeURL
	Delete             templ.SafeURL
}) {
	url.Page = BuildURL(fmt.Sprintf("/press/%d", pressNumber))
	url.ActiveTools = BuildURL(fmt.Sprintf("/press/%d/active-tools", pressNumber))
	url.MetalSheets = BuildURL(fmt.Sprintf("/press/%d/metal-sheets", pressNumber))
	url.Cycles = BuildURL(fmt.Sprintf("/press/%d/cycles", pressNumber))
	url.Notes = BuildURL(fmt.Sprintf("/press/%d/notes", pressNumber))
	url.PressRegenerations = BuildURL(fmt.Sprintf("/press/%d/regenerations", pressNumber))
	url.CycleSummaryPDF = BuildURL(fmt.Sprintf("/press/%d/cycle-summary-pdf", pressNumber))
	url.Delete = BuildURL(fmt.Sprintf("/press/%d", pressNumber))

	return url
}

func UrlPressReplaceTool(pn shared.PressNumber, p shared.Slot) templ.SafeURL {
	return BuildURLWithParams(fmt.Sprintf("/press/%d/replace-tool", pn), map[string]string{
		"tool_id":  fmt.Sprintf("%d", pn),
		"position": fmt.Sprintf("%d", p),
	})
}

// -----------------------------------------------------------------------------
// Press Regeneration URLs
// -----------------------------------------------------------------------------

// UrlPressRegeneration constructs press regeneration URLs
func UrlPressRegeneration(press shared.PressNumber, pressRegenerationID shared.EntityID) (url struct {
	Page   templ.SafeURL
	Post   templ.SafeURL
	Delete templ.SafeURL
}) {
	params := map[string]string{
		"id": fmt.Sprintf("%d", pressRegenerationID),
	}

	url.Page = BuildURL(fmt.Sprintf("/press-regeneration/%d", press))
	url.Post = url.Page
	url.Delete = BuildURLWithParams(fmt.Sprintf("/press-regeneration/%d/delete", press), params)

	return url
}

// -----------------------------------------------------------------------------
// Dialog URLs
// -----------------------------------------------------------------------------

func UrlDialogEditToolRegeneration(toolRegenerationID, toolID shared.EntityID) (url struct {
	Get, Post, Put templ.SafeURL
}) {
	url.Get = BuildURLWithParams(
		"/dialog/edit-tool-regeneration",
		map[string]string{
			"tool_id": toolID.String(),
		},
	)
	url.Post = url.Get
	url.Put = BuildURLWithParams(
		"/dialog/edit-tool-regeneration",
		map[string]string{
			"id": toolRegenerationID.String(),
		},
	)

	return url
}

func UrlDialogEditCycle(
	cycleID shared.EntityID, toolID shared.EntityID, toolChangeMode bool,
) (url struct {
	Get  templ.SafeURL
	Post templ.SafeURL
	Put  templ.SafeURL
}) {
	params := map[string]string{}
	if cycleID != 0 {
		params["id"] = fmt.Sprintf("%d", cycleID)
	}
	if toolID != 0 {
		params["tool_id"] = fmt.Sprintf("%d", toolID)
	}
	if toolChangeMode {
		params["tool_change_mode"] = "true"
	}

	url.Get = BuildURLWithParams("/dialog/edit-cycle", params)
	url.Post = BuildURL("/dialog/edit-cycle")
	url.Put = BuildURL("/dialog/edit-cycle")

	return url
}

func UrlDialogEditPress(pressID shared.PressNumber) (url struct {
	Get, Post, Put templ.SafeURL
}) {
	if pressID > -1 {
		url.Get = BuildURLWithParams("/dialog/edit-press", map[string]string{
			"id": pressID.String(),
		})
	} else {
		url.Get = BuildURL("/dialog/edit-press")
	}

	url.Post = url.Get
	url.Put = BuildURLWithParams(
		"/dialog/edit-press", map[string]string{
			"id": pressID.String(),
		},
	)

	return url
}

// -----------------------------------------------------------------------------
// TODO: Remove all of this dialog URL builders here
// -----------------------------------------------------------------------------

// UrlDialogs constructs dialog URLs
func UrlDialogs() (url struct {
	EditTool              func(toolID shared.EntityID) templ.SafeURL
	EditCassette          func(cassetteID shared.EntityID) templ.SafeURL
	EditMetalSheet        func(metalSheetID shared.EntityID, toolID shared.EntityID, position shared.Slot) templ.SafeURL
	EditNote              func(noteID shared.EntityID, linked string) templ.SafeURL
	EditPressRegeneration func(pressRegenerationID shared.EntityID) templ.SafeURL
}) {
	url.EditTool = urlEditToolDialog
	url.EditCassette = urlEditCassetteDialog
	url.EditMetalSheet = urlEditMetalSheetDialog
	url.EditNote = urlEditNoteDialog
	url.EditPressRegeneration = urlEditPressRegenerationDialog

	return url
}

// urlEditToolDialog constructs edit tool dialog URL
func urlEditToolDialog(toolID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if toolID != 0 {
		params["id"] = fmt.Sprintf("%d", toolID)
	}

	return BuildURLWithParams("/dialog/edit-tool", params)
}

func urlEditCassetteDialog(cassetteID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if cassetteID != 0 {
		params["id"] = fmt.Sprintf("%d", cassetteID)
	}

	return BuildURLWithParams("/dialog/edit-cassette", params)
}

// urlEditMetalSheetDialog constructs edit metal sheet dialog URL
func urlEditMetalSheetDialog(metalSheetID shared.EntityID, toolID shared.EntityID, position shared.Slot) templ.SafeURL {
	params := map[string]string{}
	if metalSheetID != 0 {
		params["id"] = fmt.Sprintf("%d", metalSheetID)
		params["position"] = fmt.Sprintf("%d", position)
	}
	if toolID != 0 {
		params["tool_id"] = fmt.Sprintf("%d", toolID)
	}

	return BuildURLWithParams("/dialog/edit-metal-sheet", params)
}

// urlEditNoteDialog constructs edit note dialog URL
func urlEditNoteDialog(noteID shared.EntityID, linked string) templ.SafeURL {
	params := map[string]string{
		"linked": linked,
	}
	if noteID != 0 {
		params["id"] = fmt.Sprintf("%d", noteID)
	}

	return BuildURLWithParams("/dialog/edit-note", params)
}

// urlEditPressRegenerationDialog constructs edit press regeneration dialog URL
func urlEditPressRegenerationDialog(pressRegenerationID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if pressRegenerationID != 0 {
		params["id"] = fmt.Sprintf("%d", pressRegenerationID)
	}

	return BuildURLWithParams("/dialog/edit-press-regeneration", params)
}
