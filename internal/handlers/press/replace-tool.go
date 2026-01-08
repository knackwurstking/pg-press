package press

import (
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func ReplaceTool(c echo.Context) *echo.HTTPError {
	// TODO: Replace tool and trigger a section reload ("reload-active-tools")
	// Get query values: "position", "tool_id"
	// Get form values: "tool_id" - this is the new tool to use for position

	urlb.SetHXTrigger(c, "reload-active-tools")

	return nil
}
