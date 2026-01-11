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

// UmbauPage constructs umbau URL
func UmbauPage(press shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/umbau/%d", press))
}

// -----------------------------------------------------------------------------
// Trouble Reports URLs
// -----------------------------------------------------------------------------

// TroubleReportsPage constructs trouble reports page URL
func TroubleReportsPage() templ.SafeURL {
	return BuildURL("/trouble-reports")
}

// TroubleReportsSharePDF constructs trouble reports share PDF URL
func TroubleReportsSharePDF(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/share-pdf", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsAttachment constructs trouble reports attachment URL
func TroubleReportsAttachment(trID, aID shared.EntityID, modificationTime int64) templ.SafeURL {
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
	return BuildURLWithParams("/trouble-reports/attachment", params)
}

// TroubleReportsModifications constructs trouble reports modifications URL
func TroubleReportsModifications(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams(fmt.Sprintf("/trouble-reports/modifications/%d", trID), map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsData constructs trouble reports data URL
func TroubleReportsData(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/data", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsAttachmentsPreview constructs trouble reports attachments preview URL
func TroubleReportsAttachmentsPreview(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/attachments-preview", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// TroubleReportsRollback constructs trouble reports rollback URL
func TroubleReportsRollback(trID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/trouble-reports/rollback", map[string]string{
		"id": fmt.Sprintf("%d", trID),
	})
}

// -----------------------------------------------------------------------------
// Tools URLs
// -----------------------------------------------------------------------------

// ToolsPage constructs tools page URL
func ToolsPage() templ.SafeURL {
	return BuildURL("/tools")
}

// ToolsDelete constructs tools delete URL
func ToolsDelete(toolID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/tools/delete", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

// ToolsMarkDead constructs tools mark dead URL
func ToolsMarkDead(toolID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/tools/mark-dead", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

// ToolsSectionPress constructs tools section press URL
func ToolsSectionPress() templ.SafeURL {
	return BuildURL("/tools/section/press")
}

// ToolsSectionTools constructs tools section tools URL
func ToolsSectionTools() templ.SafeURL {
	return BuildURL("/tools/section/tools")
}

// ToolsAdminOverlapping constructs admin overlapping tools URL
func ToolsAdminOverlapping() templ.SafeURL {
	return BuildURL("/tools/admin/overlapping-tools")
}

// -----------------------------------------------------------------------------
// Tool URLs
// -----------------------------------------------------------------------------

// ToolPage constructs tool page URL
func ToolPage(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d", toolID))
}

// ToolDeleteRegeneration constructs tool delete regeneration URL
func ToolDeleteRegeneration(toolID, toolRegenerationID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if toolRegenerationID != 0 {
		params["id"] = fmt.Sprintf("%d", toolRegenerationID)
	}
	return BuildURLWithParams(fmt.Sprintf("/tool/%d/delete-regeneration", toolID), params)
}

// ToolRegenerationEdit constructs tool regeneration edit URL
func ToolRegenerationEdit(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/regeneration-edit", toolID))
}

// ToolRegenerationDisplay constructs tool regeneration display URL
func ToolRegenerationDisplay(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/regeneration-display", toolID))
}

// ToolRegeneration constructs tool regeneration URL
func ToolRegeneration(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/regeneration", toolID))
}

// ToolNotes constructs tool notes URL
func ToolNotes(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/notes", toolID))
}

// ToolMetalSheets constructs tool metal sheets URL
func ToolMetalSheets(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/metal-sheets", toolID))
}

// ToolCycles constructs tool cycles URL
func ToolCycles(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/cycles", toolID))
}

// ToolTotalCycles constructs tool total cycles URL
func ToolTotalCycles(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/total-cycles", toolID))
}

// ToolCycleDelete constructs tool cycle delete URL
func ToolCycleDelete(cycleID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if cycleID != 0 {
		params["id"] = fmt.Sprintf("%d", cycleID)
	}
	return BuildURLWithParams("/tool/cycle/delete", params)
}

// ToolBind constructs tool bind URL
func ToolBind(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/bind", toolID))
}

// ToolUnbind constructs tool unbind URL
func ToolUnbind(toolID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/tool/%d/unbind", toolID))
}

// -----------------------------------------------------------------------------
// Press URLs
// -----------------------------------------------------------------------------

// PressPage constructs press page URL
func PressPage(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d", pressNumber))
}

// PressActiveTools constructs press active tools URL
func PressActiveTools(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/active-tools", pressNumber))
}

// PressMetalSheets constructs press metal sheets URL
func PressMetalSheets(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/metal-sheets", pressNumber))
}

// PressCycles constructs press cycles URL
func PressCycles(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/cycles", pressNumber))
}

// PressNotes constructs press notes URL
func PressNotes(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/notes", pressNumber))
}

// PressRegenerations constructs press regenerations URL
func PressRegenerations(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/regenerations", pressNumber))
}

// PressCycleSummaryPDF constructs press cycle summary PDF URL
func PressCycleSummaryPDF(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/cycle-summary-pdf", pressNumber))
}

// PressDelete constructs press delete URL
func PressDelete(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d", pressNumber))
}

// PressReplaceTool constructs press replace tool URL
func PressReplaceTool(pn shared.PressNumber, p shared.Slot) templ.SafeURL {
	return BuildURLWithParams(fmt.Sprintf("/press/%d/replace-tool", pn), map[string]string{
		"tool_id":  fmt.Sprintf("%d", pn),
		"position": fmt.Sprintf("%d", p),
	})
}

// -----------------------------------------------------------------------------
// Press Regeneration URLs
// -----------------------------------------------------------------------------

// PressRegenerationPage constructs press regeneration page URL
func PressRegenerationPage(press shared.PressNumber, pressRegenerationID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press-regeneration/%d", press))
}

// PressRegenerationDelete constructs press regeneration delete URL
func PressRegenerationDelete(press shared.PressNumber, pressRegenerationID shared.EntityID) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", pressRegenerationID),
	}
	return BuildURLWithParams(fmt.Sprintf("/press-regeneration/%d/delete", press), params)
}

// -----------------------------------------------------------------------------
// Dialog URLs
// -----------------------------------------------------------------------------

// DialogEditToolRegenerationGet constructs edit tool regeneration dialog GET URL
func DialogEditToolRegenerationGet(toolID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams(
		"/dialog/edit-tool-regeneration",
		map[string]string{
			"tool_id": toolID.String(),
		},
	)
}

// DialogEditToolRegenerationPost constructs edit tool regeneration dialog POST URL
func DialogEditToolRegenerationPost(toolID shared.EntityID) templ.SafeURL {
	return DialogEditToolRegenerationGet(toolID)
}

// DialogEditToolRegenerationPut constructs edit tool regeneration dialog PUT URL
func DialogEditToolRegenerationPut(toolRegenerationID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams(
		"/dialog/edit-tool-regeneration",
		map[string]string{
			"id": toolRegenerationID.String(),
		},
	)
}

// DialogEditCycleGet constructs edit cycle dialog GET URL
func DialogEditCycleGet(cycleID shared.EntityID, toolID shared.EntityID, toolChangeMode bool) templ.SafeURL {
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
	return BuildURLWithParams("/dialog/edit-cycle", params)
}

// DialogEditCyclePost constructs edit cycle dialog POST URL
func DialogEditCyclePost() templ.SafeURL {
	return BuildURL("/dialog/edit-cycle")
}

// DialogEditCyclePut constructs edit cycle dialog PUT URL
func DialogEditCyclePut() templ.SafeURL {
	return BuildURL("/dialog/edit-cycle")
}

// DialogEditPressGet constructs edit press dialog GET URL
func DialogEditPressGet(pressID shared.PressNumber) templ.SafeURL {
	if pressID > -1 {
		return BuildURLWithParams("/dialog/edit-press", map[string]string{
			"id": pressID.String(),
		})
	}
	return BuildURL("/dialog/edit-press")
}

// DialogEditPressPost constructs edit press dialog POST URL
func DialogEditPressPost(pressID shared.PressNumber) templ.SafeURL {
	return DialogEditPressGet(pressID)
}

// DialogEditPressPut constructs edit press dialog PUT URL
func DialogEditPressPut(pressID shared.PressNumber) templ.SafeURL {
	return BuildURLWithParams(
		"/dialog/edit-press", map[string]string{
			"id": pressID.String(),
		},
	)
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
