// TODO: Move this stuff somewhere else, or just split it up into multiple files
package helpers

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// Category: Cycles

func HXGetCycleEditDialog(toolID int64, cycleID *int64, toolChangeMode bool) templ.SafeURL {
	props := make([]string, 0)

	props = append(props, fmt.Sprintf("tool_id=%d", toolID))

	if cycleID != nil {
		props = append(props, fmt.Sprintf("id=%d", *cycleID))
	}

	props = append(props, fmt.Sprintf("tool_change_mode=%t", toolChangeMode))

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/cycle/edit?%s",
		env.ServerPathPrefix, strings.Join(props, "&"),
	))
}

func HXPostCycleEditDialog(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/cycle/edit?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPutCycleEditDialog(cycleID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/cycle/edit?id=%d",
		env.ServerPathPrefix, cycleID,
	))
}

func HXDeleteCycle(cycleID, toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/cycle/delete?id=%d&tool_id=%d",
		env.ServerPathPrefix, cycleID, toolID,
	))
}

// Category: Feeds

func HXGetFeedList() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/feed/list",
		env.ServerPathPrefix,
	))
}

// Category: Metal Sheets

func HXGetMetalSheetEditDialog(metalSheetID *int64, toolID *int64) templ.SafeURL {
	if metalSheetID == nil && toolID != nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/metal-sheets/edit&tool_id=%d",
			env.ServerPathPrefix, *toolID,
		))
	}

	if metalSheetID != nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/metal-sheets/edit?id=%d",
			env.ServerPathPrefix, *metalSheetID,
		))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit",
		env.ServerPathPrefix,
	))

}

func HXPostMetalSheetEditDialog(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPutMetalSheetEditDialog(metalSheetID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?id=%d",
		env.ServerPathPrefix, metalSheetID,
	))
}

func HXDeleteMetalSheet(metalSheetID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/delete?id=%d",
		env.ServerPathPrefix, metalSheetID,
	))
}

// Category: Notes

func HXGetNotesEditDialog(noteID *int64, linkToTables ...string) templ.SafeURL {
	var params []string

	if noteID != nil {
		params = append(params, fmt.Sprintf("id=%d", *noteID))
	}

	if len(linkToTables) > 0 {
		params = append(params, fmt.Sprintf(
			"link_to_tables=%s", strings.Join(linkToTables, ","),
		))
	}

	url := fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix)
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	return templ.SafeURL(url)
}

func HXPostNotesEditDialog(linkToTables ...string) templ.SafeURL {
	if len(linkToTables) == 0 {
		return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/notes/edit?link_to_tables=%s",
		env.ServerPathPrefix, strings.Join(linkToTables, ","),
	))
}

func HXPutNotesEditDialog(noteID int64, linkToTables ...string) templ.SafeURL {
	if len(linkToTables) == 0 {
		return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/edit?id=%d", env.ServerPathPrefix, noteID))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/notes/edit?id=%d&link_to_tables=%s",
		env.ServerPathPrefix, noteID, strings.Join(linkToTables, ","),
	))
}

func HXDeleteNote(noteID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/delete?id=%d", env.ServerPathPrefix, noteID))
}

// Category: Tool Page

func HXGetToolNotesSectionContent(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/notes?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXGetToolMetalSheetsSectionContent(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/metal-sheets?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXGetToolCyclesSectionContent(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/cycles?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

// Category: Tools Page

func HXGetToolsPagePressSectionContent() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/section/press",
		env.ServerPathPrefix,
	))
}

func HXGetToolsPageAllToolsSectionContent() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/section/tools",
		env.ServerPathPrefix,
	))
}

// Category: Tools

func HXGetToolEditDialog(toolID *int64) templ.SafeURL {
	if toolID == nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/tools/edit",
			env.ServerPathPrefix,
		))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/edit?id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPostToolEditDialog() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/edit",
		env.ServerPathPrefix,
	))
}

func HXPutToolEditDialog(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/edit?id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPatchToolMarkDead(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/mark-dead?id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXGetToolOverlappingTools() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/admin/overlapping-tools",
		env.ServerPathPrefix,
	))
}

func HXGetToolTotalCycles(toolID int64, toolPosition models.Position, input bool) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/total-cycles?tool_id=%d&tool_position=%s&input=%t",
		env.ServerPathPrefix, toolID, toolPosition, input,
	))
}

func HXGetToolStatusEdit(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/status-edit?id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPutToolStatus() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/status",
		env.ServerPathPrefix,
	))
}

func HXGetToolStatusDisplay(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/status-display?id=%d",
		env.ServerPathPrefix, toolID,
	))
}

// Category: Press Page

func HXGetPressNotesSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/press/%d/notes",
		env.ServerPathPrefix, pressNumber,
	))
}

func HXGetPressActiveToolsSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/press/%d/active-tools",
		env.ServerPathPrefix, pressNumber,
	))
}

func HXGetPressMetalSheetsSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/press/%d/metal-sheets",
		env.ServerPathPrefix, pressNumber,
	))
}

func HXGetPressCyclesSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/press/%d/cycles",
		env.ServerPathPrefix, pressNumber,
	))
}

// Category: Press

func HXGetPressCycleSummaryPDF(pressNumber models.PressNumber) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/press/%d/cycle-summary-pdf",
		env.ServerPathPrefix, pressNumber,
	))
}

// Category: Regeneration

func HXGetToolRegenerationEdit(toolID int64, regenerationID *int64) templ.SafeURL {
	if regenerationID == nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/tools/tool/%d/edit-regeneration",
			env.ServerPathPrefix, toolID,
		))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/edit-regeneration?id=%d",
		env.ServerPathPrefix, toolID, *regenerationID,
	))
}

func HXPutToolRegenerationEdit(toolID int64, regenerationID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/edit-regeneration?id=%d",
		env.ServerPathPrefix, toolID, regenerationID,
	))
}

func HXDeleteToolRegeneration(toolID int64, regenerationID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/delete-regeneration?id=%d",
		env.ServerPathPrefix, toolID, regenerationID,
	))
}

func HXPatchToolBind(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/bind",
		env.ServerPathPrefix, toolID,
	))
}

func HXPatchToolUnbind(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/unbind",
		env.ServerPathPrefix, toolID,
	))
}

// Category: Trouble Reports

func HXGetTroubleReportsData() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/data",
		env.ServerPathPrefix,
	))
}

func HXDeleteTroubleReportsData(troubleReportID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/data?id=%d",
		env.ServerPathPrefix, troubleReportID,
	))
}

func HXGetTroubleReportsAttachmentsPreview(troubleReportID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/attachments-preview?id=%d",
		env.ServerPathPrefix, troubleReportID,
	))
}

func HXPostTroubleReportsRollback(troubleReportID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/trouble-reports/rollback?id=%d",
		env.ServerPathPrefix, troubleReportID,
	))
}

// Category: Cookies

func HXGetCookies() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/profile/cookies",
		env.ServerPathPrefix,
	))
}

func HXDeleteCookies(value string) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/profile/cookies?value=%s",
		env.ServerPathPrefix, value,
	))
}

// Category: Nav

func HXWsConnectNavFeedCounter() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"connect:%s/htmx/nav/feed-counter",
		env.ServerPathPrefix,
	))
}
