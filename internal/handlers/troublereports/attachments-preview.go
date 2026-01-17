package troublereports

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/troublereports/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func GetAttachmentsPreview(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id") // Get trouble report ID
	if merr != nil {
		return merr.Echo()
	}

	tr, merr := db.GetTroubleReport(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	t := templates.AttachmentsPreview(tr.LinkedAttachments)
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Attachments Preview")
	}

	return nil
}
