package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/models"
)

func HXGetMetalSheetEditDialog(metalSheetID *models.MetalSheetID, toolID *models.ToolID) templ.SafeURL {
	params := make(map[string]string)

	if metalSheetID != nil {
		params["id"] = fmt.Sprintf("%d", *metalSheetID)
	} else if toolID != nil {
		params["tool_id"] = fmt.Sprintf("%d", *toolID)
	}

	return buildURL("/htmx/dialogs/edit-metal-sheet", params)
}

func HXPostMetalSheetEditDialog(toolID models.ToolID) templ.SafeURL {
	params := map[string]string{
		"tool_id": fmt.Sprintf("%d", toolID),
	}
	return buildURL("/htmx/dialogs/edit-metal-sheet", params)
}

func HXPutMetalSheetEditDialog(metalSheetID models.MetalSheetID) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	}
	return buildURL("/htmx/dialogs/edit-metal-sheet", params)
}

func HXDeleteMetalSheet(metalSheetID models.MetalSheetID) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	}
	return buildURL("/htmx/metal-sheets/delete", params)
}
