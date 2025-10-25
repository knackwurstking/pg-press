package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/models"
)

func HXGetToolRegenerationEdit(toolID int64, regenerationID *models.RegenerationID) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/edit-regeneration", toolID)

	if regenerationID == nil {
		return buildURL(path, nil)
	}

	params := map[string]string{
		"id": fmt.Sprintf("%d", *regenerationID),
	}
	return buildURL(path, params)
}

func HXPutToolRegenerationEdit(toolID int64, regenerationID models.RegenerationID) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/edit-regeneration", toolID)
	params := map[string]string{
		"id": fmt.Sprintf("%d", regenerationID),
	}
	return buildURL(path, params)
}

func HXDeleteToolRegeneration(toolID int64, regenerationID models.RegenerationID) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/delete-regeneration", toolID)
	params := map[string]string{
		"id": fmt.Sprintf("%d", regenerationID),
	}
	return buildURL(path, params)
}

func HXPatchToolBind(toolID int64) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/bind", toolID)
	return buildURL(path, nil)
}

func HXPatchToolUnbind(toolID int64) templ.SafeURL {
	path := fmt.Sprintf("/htmx/tools/tool/%d/unbind", toolID)
	return buildURL(path, nil)
}
