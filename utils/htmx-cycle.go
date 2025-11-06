package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/models"
)

// Cycle-related functions

func HXGetCycleEditDialog(toolID models.ToolID, cycleID *models.CycleID, toolChangeMode bool) templ.SafeURL {
	params := map[string]string{
		"tool_id":          fmt.Sprintf("%d", toolID),
		"tool_change_mode": fmt.Sprintf("%t", toolChangeMode),
	}

	if cycleID != nil {
		params["id"] = fmt.Sprintf("%d", *cycleID)
	}

	return buildURL("/htmx/dialogs/edit-cycle", params)
}

func HXPostCycleEditDialog(toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/dialogs/edit-cycle", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXPutCycleEditDialog(cycleID models.CycleID) templ.SafeURL {
	return buildURL("/htmx/dialogs/edit-cycle", map[string]string{
		"id": fmt.Sprintf("%d", cycleID),
	})
}

func HXDeleteCycle(cycleID models.CycleID, toolID models.ToolID) templ.SafeURL {
	return buildURL("/htmx/tools/cycle/delete", map[string]string{
		"id":      fmt.Sprintf("%d", cycleID),
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}
