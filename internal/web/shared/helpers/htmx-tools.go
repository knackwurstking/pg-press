// TODO: Split this stuff up into multiple files, a bit more
package helpers

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/pkg/models"
)

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
