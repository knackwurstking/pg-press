package utils

import (
	"fmt"

	"github.com/a-h/templ"
)

func HXGetMetalSheetEditDialog(metalSheetID *int64, toolID *int64) templ.SafeURL {
	params := make(map[string]string)

	if metalSheetID != nil {
		params["id"] = fmt.Sprintf("%d", *metalSheetID)
	} else if toolID != nil {
		params["tool_id"] = fmt.Sprintf("%d", *toolID)
	}

	return buildURL("/htmx/metal-sheets/edit", params)
}

func HXPostMetalSheetEditDialog(toolID int64) templ.SafeURL {
	params := map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	}
	return buildURL("/htmx/metal-sheets/edit", params)
}

func HXPutMetalSheetEditDialog(metalSheetID int64) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	}
	return buildURL("/htmx/metal-sheets/edit", params)
}

func HXDeleteMetalSheet(metalSheetID int64) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	}
	return buildURL("/htmx/metal-sheets/delete", params)
}
