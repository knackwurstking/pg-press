package dialogs

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetEditToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil && merr.Code() != http.StatusNotFound {
		return merr.Echo()
	}

	if id > 0 {
		// TODO: If ID is valid, Create `EditToolRegenerationDialog` with just the tool ID
	}

	// TODO: Else, create `EditToolRegenerationDialog` for an existing tool regeneration
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	// TODO: ...
}

func PutToolRegeneration(c echo.Context) *echo.HTTPError {
	// TODO: ...
}
