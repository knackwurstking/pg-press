package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// Press constructs press page URL
func Press(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d", pressNumber))
}

// PressActiveTools constructs press active tools URL
func PressActiveTools(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/active-tools", pressNumber))
}

// PressMetalSheets constructs press metal sheets URL
func PressMetalSheets(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/metal-sheets", pressNumber))
}

// PressCycles constructs press cycles URL
func PressCycles(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/cycles", pressNumber))
}

// PressNotes constructs press notes URL
func PressNotes(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/notes", pressNumber))
}

// PressRegenerations constructs press regenerations URL
func PressRegenerations(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/regenerations", pressNumber))
}

// PressCycleSummaryPDF constructs press cycle summary PDF URL
func PressCycleSummaryPDF(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d/cycle-summary-pdf", pressNumber))
}

// PressDelete constructs press delete URL
func PressDelete(pressNumber shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press/%d", pressNumber))
}

// PressReplaceTool constructs press replace tool URL
func PressReplaceTool(pn shared.PressNumber, p shared.Slot) templ.SafeURL {
	return BuildURLWithParams(fmt.Sprintf("/press/%d/replace-tool", pn), map[string]string{
		"position": fmt.Sprintf("%d", p),
	})
}
