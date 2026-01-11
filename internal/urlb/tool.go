package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// Tool constructs tool page URL
func Tool(toolID shared.EntityID) templ.SafeURL {
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
