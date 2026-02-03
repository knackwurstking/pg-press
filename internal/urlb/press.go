package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// Press constructs press page URL
func Press(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d", pressID))
}

// PressActiveTools constructs press active tools URL
func PressActiveTools(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/active-tools", pressID))
}

// PressMetalSheets constructs press metal sheets URL
// TODO: Continue here...
func PressMetalSheets(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/metal-sheets", pressID))
}

// PressCycles constructs press cycles URL
func PressCycles(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/cycles", pressID))
}

// PressNotes constructs press notes URL
func PressNotes(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/notes", pressID))
}

// PressCycleSummaryPDF constructs press cycle summary PDF URL
func PressCycleSummaryPDF(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/cycle-summary-pdf", pressID))
}

// PressDelete constructs press delete URL
func PressDelete(pressID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d", pressID))
}

// PressReplaceTool constructs press replace tool URL
func PressReplaceTool(pressID shared.EntityID, p shared.Slot) templ.SafeURL {
	return BuildURLWithParams(fmt.Sprintf("/press/%d/replace-tool", pressID), map[string]string{
		"position": fmt.Sprintf("%d", p),
	})
}
