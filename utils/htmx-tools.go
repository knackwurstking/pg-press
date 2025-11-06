package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/models"
)

// Tool sections

func HXGetToolNotesSectionContent(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/notes", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXGetToolMetalSheetsSectionContent(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/metal-sheets", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXGetToolCyclesSectionContent(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/cycles", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HxGetToolsPageToolsSection() templ.SafeURL {
	return buildURL("/htmx/tools/section/tools", nil)
}

// Tool edit dialog

func HXGetToolEditDialog(toolID *models.ToolID) templ.SafeURL {
	if toolID == nil {
		return buildURL("/htmx/dialogs/edit-tool", nil)
	}

	return buildURL("/htmx/dialogs/edit-tool", map[string]string{
		"id": fmt.Sprintf("%d", *toolID),
	})
}

func HXPostToolEditDialog() templ.SafeURL {
	return buildURL("/htmx/dialogs/edit-tool", nil)
}

func HXPutToolEditDialog(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/dialogs/edit-tool", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

// Tool operations

func HXPatchToolMarkDead(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/mark-dead", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

func HxGetToolsPageAdminTools() templ.SafeURL {
	return buildURL("/htmx/tools/admin/overlapping-tools", nil)
}

func HXGetToolTotalCycles(toolID models.ToolID, toolPosition models.Position, input bool) templ.SafeURL {
	return buildURL("/htmx/tools/total-cycles", map[string]string{
		"tool_id":       fmt.Sprintf("%d", toolID),
		"tool_position": string(toolPosition),
		"input":         fmt.Sprintf("%t", input),
	})
}

// Tool status

func HXGetToolStatusEdit(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/status-edit", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

func HXPutToolStatus() templ.SafeURL {
	return buildURL("/htmx/tools/status", nil)
}

func HXGetToolStatusDisplay(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/status-display", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}
