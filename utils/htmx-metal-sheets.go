package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/env"
)

func HXGetMetalSheetEditDialog(metalSheetID *int64, toolID *int64) templ.SafeURL {
	if metalSheetID == nil && toolID != nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/metal-sheets/edit?tool_id=%d",
			env.ServerPathPrefix, *toolID,
		))
	}

	if metalSheetID != nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/metal-sheets/edit?id=%d",
			env.ServerPathPrefix, *metalSheetID,
		))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit",
		env.ServerPathPrefix,
	))

}

func HXPostMetalSheetEditDialog(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?tool_id=%d",
		env.ServerPathPrefix, toolID,
	))
}

func HXPutMetalSheetEditDialog(metalSheetID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?id=%d",
		env.ServerPathPrefix, metalSheetID,
	))
}

func HXDeleteMetalSheet(metalSheetID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/metal-sheets/delete?id=%d",
		env.ServerPathPrefix, metalSheetID,
	))
}
