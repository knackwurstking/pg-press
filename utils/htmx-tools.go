package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/models"
)

// Tool sections

func HXGetToolNotesSectionContent(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/notes", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXGetToolMetalSheetsSectionContent(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/metal-sheets", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXGetToolCyclesSectionContent(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/cycles", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXGetToolsPageAllToolsSectionContent() templ.SafeURL {
	return buildURL("/htmx/tools/section/tools", nil)
}

// Tool edit dialog

func HXGetToolEditDialog(toolID *int64) templ.SafeURL {
	if toolID == nil {
		return buildURL("/htmx/tools/edit", nil)
	}

	return buildURL("/htmx/tools/edit", map[string]string{
		"id": fmt.Sprintf("%d", *toolID),
	})
}

func HXPostToolEditDialog() templ.SafeURL {
	return buildURL("/htmx/tools/edit", nil)
}

func HXPutToolEditDialog(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/edit", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

// Tool operations

func HXPatchToolMarkDead(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/mark-dead", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

func HXGetToolOverlappingTools() templ.SafeURL {
	return buildURL("/htmx/tools/admin/overlapping-tools", nil)
}

func HXGetToolTotalCycles(toolID int64, toolPosition models.Position, input bool) templ.SafeURL {
	return buildURL("/htmx/tools/total-cycles", map[string]string{
		"tool_id":       fmt.Sprintf("%d", toolID),
		"tool_position": string(toolPosition),
		"input":         fmt.Sprintf("%t", input),
	})
}

// Tool status

func HXGetToolStatusEdit(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/status-edit", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

func HXPutToolStatus() templ.SafeURL {
	return buildURL("/htmx/tools/status", nil)
}

func HXGetToolStatusDisplay(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/status-display", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}
