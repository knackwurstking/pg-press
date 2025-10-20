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
		env.ServerPathPrefix, *toolID,
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
