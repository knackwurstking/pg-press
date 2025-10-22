package utils

import (
	"fmt"

	"github.com/a-h/templ"
)

// Cycle-related functions

func HXGetCycleEditDialog(toolID int64, cycleID *int64, toolChangeMode bool) templ.SafeURL {
	params := map[string]string{
		"tool_id":          fmt.Sprintf("%d", toolID),
		"tool_change_mode": fmt.Sprintf("%t", toolChangeMode),
	}

	if cycleID != nil {
		params["id"] = fmt.Sprintf("%d", *cycleID)
	}

	return buildURL("/htmx/tools/cycle/edit", params)
}

func HXPostCycleEditDialog(toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/cycle/edit", map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}

func HXPutCycleEditDialog(cycleID int64) templ.SafeURL {
	return buildURL("/htmx/tools/cycle/edit", map[string]string{
		"id": fmt.Sprintf("%d", cycleID),
	})
}

func HXDeleteCycle(cycleID, toolID int64) templ.SafeURL {
	return buildURL("/htmx/tools/cycle/delete", map[string]string{
		"id":      fmt.Sprintf("%d", cycleID),
		"tool_id": fmt.Sprintf("%d", toolID),
	})
}
