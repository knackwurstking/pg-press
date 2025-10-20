package helpers

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// Category: Cycles

func HXGetCycleEdit(toolID int64, cycleID *int64, toolChangeMode bool) templ.SafeURL {
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

func HXPostCycleEdit(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/cycle/edit?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPutCycleEdit(cycleID int64) templ.SafeURL {
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

func HXGetMetalSheetEdit(metalSheetID *int64, toolID *int64) templ.SafeURL {
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

func HXPostMetalSheetEdit(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPutMetalSheetEdit(metalSheetID int64) templ.SafeURL {
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

func HXGetNotesEdit(noteID *int64, linkToTables ...string) templ.SafeURL {
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

func HXPostNotesEdit(linkToTables ...string) templ.SafeURL {
	if len(linkToTables) == 0 {
		return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/notes/edit?link_to_tables=%s",
		env.ServerPathPrefix, strings.Join(linkToTables, ","),
	))
}

func HXPutNotesEdit(noteID int64, linkToTables ...string) templ.SafeURL {
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

// Category: Press

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

func HXGetPressCycleSummaryPDF(pressNumber models.PressNumber) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/press/%d/cycle-summary-pdf",
		env.ServerPathPrefix, pressNumber,
	))
}

// Category: Tool

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

// Patch
// "%s/htmx/tools/tool/%d/bind"
func HXPatchToolBind(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/bind",
		env.ServerPathPrefix, toolID,
	))
}

// Patch
// "%s/htmx/tools/tool/%d/unbind"
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
