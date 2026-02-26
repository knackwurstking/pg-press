package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	if id > 0 {
		tr, merr := db.GetToolRegeneration(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
		t := EditToolRegenerationDialog(tr)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolRegenerationDialog")
		}
		return nil
	}

	id, merr = utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	t := NewToolRegenerationDialog(shared.EntityID(id))
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}
	return nil
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}

	formData, verr := parseEditToolRegenerationForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	merr = db.AddToolRegeneration(&shared.ToolRegeneration{
		ToolID: shared.EntityID(id),
		Start:  formData.Start,
		Stop:   formData.Stop,
	})
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return nil
}

func PutToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	tr, merr := db.GetToolRegeneration(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	formData, verr := parseEditToolRegenerationForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	tr.Start = formData.Start
	tr.Stop = formData.Stop
	merr = db.UpdateToolRegeneration(tr)
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return nil
}

type editToolRegenerationForm struct {
	Start shared.UnixMilli
	Stop  shared.UnixMilli
}

func parseEditToolRegenerationForm(c echo.Context) (*editToolRegenerationForm, *errors.ValidationError) {
	// Parse start and stop dates from HTML input fields (type "date")
	v, err := utils.SanitizeInt64(c.FormValue("start"))
	if err != nil {
		return nil, errors.NewValidationError("invalid start date")
	}
	start := shared.UnixMilli(v)

	v, err = utils.SanitizeInt64(c.FormValue("stop"))
	if err != nil {
		return nil, errors.NewValidationError("invalid stop date")
	}
	stop := shared.UnixMilli(v)

	return &editToolRegenerationForm{Start: start, Stop: stop}, nil
}
