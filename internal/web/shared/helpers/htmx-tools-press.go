package helpers

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func HXGetToolsPagePressSectionContent() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/section/press",
		env.ServerPathPrefix,
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
