package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func DeleteRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := urlb.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	regeneration, merr := db.GetToolRegeneration(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	merr = db.DeleteToolRegeneration(regeneration.ID)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "reload-cycles")

	return nil
}
