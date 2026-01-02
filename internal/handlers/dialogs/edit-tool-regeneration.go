package dialogs

import (
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetEditToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
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

	id, merr = shared.ParseQueryInt64(c, "tool_id")
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
	id, merr := shared.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}

	formData, merr := GetEditToolRegenerationForm(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = db.AddToolRegeneration(&shared.ToolRegeneration{
		ToolID: shared.EntityID(id),
		Start:  formData.Start,
		Stop:   formData.Stop,
	})
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "reload-tool-regenerations")

	return nil
}

func PutToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	formData, merr := GetEditToolRegenerationForm(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = db.UpdateToolRegeneration(&shared.ToolRegeneration{
		ID:     shared.EntityID(id),
		ToolID: shared.EntityID(id),
		Start:  formData.Start,
		Stop:   formData.Stop,
	})
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "reload-tool-regenerations")

	return nil
}

type EditToolRegenerationForm struct {
	Start shared.UnixMilli
	Stop  shared.UnixMilli
}

func GetEditToolRegenerationForm(c echo.Context) (data EditToolRegenerationForm, merr *errors.MasterError) {
	// Parse start and stop dates from HTML input fields (type "date")
	vStart := c.FormValue("start")
	vStop := c.FormValue("stop")

	if vStart == "" || vStop == "" {
		return data, errors.NewValidationError("missing start or stop").MasterError()
	}

	startInt, err := strconv.ParseInt(vStart, 10, 64)
	if err != nil {
		return data, errors.NewValidationError("invalid start date").MasterError()
	}
	stopInt, err := strconv.ParseInt(vStop, 10, 64)
	if err != nil {
		return data, errors.NewValidationError("invalid stop date").MasterError()
	}

	data.Start = shared.UnixMilli(startInt)
	data.Stop = shared.UnixMilli(stopInt)

	return data, nil
}
