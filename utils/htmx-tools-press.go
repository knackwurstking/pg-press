package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/models"
)

func HXGetToolsPagePressSectionContent() templ.SafeURL {
	return buildURL("/htmx/tools/section/press", nil)
}

func HXGetPressNotesSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return buildURL(fmt.Sprintf("/htmx/tools/press/%d/notes", pressNumber), nil)
}

func HXGetPressActiveToolsSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return buildURL(fmt.Sprintf("/htmx/tools/press/%d/active-tools", pressNumber), nil)
}

func HXGetPressMetalSheetsSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return buildURL(fmt.Sprintf("/htmx/tools/press/%d/metal-sheets", pressNumber), nil)
}

func HXGetPressCyclesSectionContent(pressNumber models.PressNumber) templ.SafeURL {
	return buildURL(fmt.Sprintf("/htmx/tools/press/%d/cycles", pressNumber), nil)
}

func HXGetPressCycleSummaryPDF(pressNumber models.PressNumber) templ.SafeURL {
	return buildURL(fmt.Sprintf("/htmx/tools/press/%d/cycle-summary-pdf", pressNumber), nil)
}
