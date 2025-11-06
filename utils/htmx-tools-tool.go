package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/models"
)

func HXGetToolRegenerationEdit(regenerationID models.RegenerationID) templ.SafeURL {
	return buildURL(
		fmt.Sprintf("/htmx/dialogs/edit-regeneration?id=%d", regenerationID),
		nil,
	)
}

func HXPutToolRegenerationEdit(toolID models.ToolID, regenerationID models.RegenerationID) templ.SafeURL {
	return buildURL(
		fmt.Sprintf("/htmx/dialogs/edit-regeneration?id=%d", regenerationID),
		nil,
	)
}

func HXDeleteToolRegeneration(toolID models.ToolID, regenerationID models.RegenerationID) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/delete-regeneration", toolID)
	params := map[string]string{
		"id": fmt.Sprintf("%d", regenerationID),
	}
	return buildURL(path, params)
}

func HXPatchToolBind(toolID models.ToolID) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/bind", toolID)
	return buildURL(path, nil)
}

func HXPatchToolUnbind(toolID models.ToolID) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/unbind", toolID)
	return buildURL(path, nil)
}
