package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// Tools constructs tools page URL
func Tools() templ.SafeURL {
	return BuildURL("/tools")
}

// ToolsDelete constructs tools delete URL
func ToolsDelete(toolID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/tools/delete", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

// ToolsMarkDead constructs tools mark dead URL
func ToolsMarkDead(toolID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/tools/mark-dead", map[string]string{
		"id": fmt.Sprintf("%d", toolID),
	})
}

// ToolsSectionPress constructs tools section press URL
func ToolsSectionPress() templ.SafeURL {
	return BuildURL("/tools/section/press")
}

// ToolsSectionTools constructs tools section tools URL
func ToolsSectionTools() templ.SafeURL {
	return BuildURL("/tools/section/tools")
}

// ToolsAdminOverlapping constructs admin overlapping tools URL
func ToolsAdminOverlapping() templ.SafeURL {
	return BuildURL("/tools/admin/overlapping-tools")
}
