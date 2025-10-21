package utils

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/env"
)

func HXGetToolRegenerationEdit(toolID int64, regenerationID *int64) templ.SafeURL {
	if regenerationID == nil {
		return templ.SafeURL(fmt.Sprintf(
			"%s/htmx/tools/tool/%d/edit-regeneration",
			env.ServerPathPrefix, toolID,
		))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/edit-regeneration?id=%d",
		env.ServerPathPrefix, toolID, *regenerationID,
	))
}

func HXPutToolRegenerationEdit(toolID int64, regenerationID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/edit-regeneration?id=%d",
		env.ServerPathPrefix, toolID, regenerationID,
	))
}

func HXDeleteToolRegeneration(toolID int64, regenerationID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/delete-regeneration?id=%d",
		env.ServerPathPrefix, toolID, regenerationID,
	))
}

func HXPatchToolBind(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/bind",
		env.ServerPathPrefix, toolID,
	))
}

func HXPatchToolUnbind(toolID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/tools/tool/%d/unbind",
		env.ServerPathPrefix, toolID,
	))
}
