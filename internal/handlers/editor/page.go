package editor

import (
	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func Page(c echo.Context) *echo.HTTPError {
	// TODO: Query: "type", "id", "return_url"

	// TODO: Render editor page
	return echo.NewHTTPError(501, "Not implemented")
}

func getQueryEditorType(c echo.Context) shared.EditorType {
	// TODO: ...
}

func getQueryID(c echo.Context) shared.EntityID {
	// TODO: ...
}

func getQueryReturnURL(c echo.Context) templ.SafeURL {
	// TODO: ...
}
