package dialogs

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditToolRegeneration(c echo.Context) *echo.HTTPError {
	if c.QueryParam("id") != "" {
		trIDQuery, herr := utils.GetQueryInt64(c, "id")
		if herr != nil {
			return herr.Echo()
		}

		tr, herr := db.GetToolRegeneration(shared.EntityID(trIDQuery))
		if herr != nil {
			return herr.Echo()
		}

		t := EditToolRegenerationDialog(tr.ID, ToolRegenerationDialogProps{
			ToolRegenerationFormData: ToolRegenerationFormData{
				Start: tr.Start,
				Stop:  tr.Stop,
			},
			Open: true,
			OOB:  true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolRegenerationDialog")
		}

		return nil
	}

	if c.QueryParam("tool_id") == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing tool or tool regeneration ID")
	}

	id, merr := utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}

	t := NewToolRegenerationDialog(ToolRegenerationFormData{
		ToolID: shared.EntityID(id),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	return nil
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	if id, _ := utils.GetQueryInt64(c, "id"); id != 0 {
		return updateToolRegeneration(c, shared.EntityID(id))
	}

	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderNewToolRegenerationDialog(c, true, data, ierrs...)
	}

	tr := shared.NewToolRegeneration(data.ToolID, data.Start, data.Stop)
	if merr := db.AddToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to create tool regeneration: %v", merr))
		return reRenderNewToolRegenerationDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderNewToolRegenerationDialog(c, false, data)
}

func updateToolRegeneration(c echo.Context, trID shared.EntityID) *echo.HTTPError {
	tr, merr := db.GetToolRegeneration(trID)
	if merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to load tool regeneration with ID %d: %v", trID, merr))
		return reRenderEditToolRegenerationDialog(c, trID, true, ToolRegenerationFormData{}, ierr)
	}

	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderEditToolRegenerationDialog(c, trID, true, data, ierrs...)
	}
	tr.Start = data.Start
	tr.Stop = data.Stop

	if merr := db.UpdateToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to update tool regeneration: %v", merr))
		return reRenderEditToolRegenerationDialog(c, trID, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderEditToolRegenerationDialog(c, trID, false, data)
}

type ToolRegenerationFormData struct {
	ToolID shared.EntityID
	Start  shared.UnixMilli
	Stop   shared.UnixMilli
}

func parseToolRegenerationForm(c echo.Context) (data ToolRegenerationFormData, ierrs []*errors.InputError) {
	data.ToolID = shared.EntityID(c.FormValue("tool_id"))

	startTime, err := utils.ParseDate(c.FormValue("start"))
	if err != nil {
		ierr := errors.NewInputError("start", fmt.Sprintf("invalid start date: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.Start = startTime

	stopTime, err := utils.ParseDate(c.FormValue("stop"))
	if err != nil {
		ierr := errors.NewInputError("stop", fmt.Sprintf("invalid stop date: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.Stop = stopTime

	return
}

func reRenderNewToolRegenerationDialog(c echo.Context, open bool, formData ToolRegenerationFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := NewToolRegenerationDialog(ToolRegenerationFormData{
		ToolID: formData.ToolID,
		Start:  formData.Start,
		Stop:   formData.Stop,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}

func reRenderEditToolRegenerationDialog(c echo.Context, trID shared.EntityID, open bool, formData ToolRegenerationFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := EditToolRegenerationDialog(trID, ToolRegenerationDialogProps{
		ToolRegenerationFormData: formData,
		Open:                     open,
		OOB:                      true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}

	if c.QueryParam("id") != "" {
		trID, herr := utils.GetQueryInt64(c, "id")
		if herr != nil {
			return herr.Echo()
		}

		tr, herr := db.GetToolRegeneration(shared.EntityID(trID))
		if herr != nil {
			ierr := errors.NewInputError("", fmt.Sprintf("failed to load regeneration: %v", herr))
			return reRenderEditToolRegeneration(c, shared.EntityID(trID), true, ToolRegenerationFormData{}, ierr)
		}

		t := EditToolRegenerationDialog(tr.ID, ToolRegenerationDialogProps{
			ToolRegenerationFormData: ToolRegenerationFormData{
				Start: tr.Start,
				Stop:  tr.Stop,
			},
			Open: true,
			OOB:  true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolRegenerationDialog")
		}

		return nil
	}

	t := NewToolRegenerationDialog(ToolRegenerationDialogProps{
		ToolRegenerationFormData: ToolRegenerationFormData{
			ToolID: shared.EntityID(toolID),
		},
		Open: true,
		OOB:  true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	return nil
}

		tr, herr := db.GetToolRegeneration(shared.EntityID(id))
		if herr != nil {
			return herr.Echo()
		}

		t := EditToolRegenerationDialog(tr.ID, ToolRegenerationDialogProps{
			ToolRegenerationFormData: ToolRegenerationFormData{
				Start: tr.Start,
				Stop:  tr.Stop,
			},
			Open: true,
			OOB:  true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolRegenerationDialog")
		}

		return nil
	}

	if c.QueryParam("tool_id") == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing tool or regen ID")
	}

	id, merr := utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}

	t := NewToolRegenerationDialog(ToolRegenerationDialogProps{
		ToolRegenerationFormData: ToolRegenerationFormData{
			ToolID: shared.EntityID(id),
		},
		Open: true,
		OOB:  true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	return nil
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderNewToolRegenerationDialog(c, true, data, ierrs...)
	}

	tr := &shared.ToolRegeneration{
		ToolID: data.ToolID,
		Start:  data.Start,
		Stop:   data.Stop,
	}
	if verr := tr.Validate(); verr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("validation error: %v", verr))
		return reRenderNewToolRegenerationDialog(c, true, data, ierr)
	}

	if merr := db.AddToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to create tool regeneration: %v", merr))
		return reRenderNewToolRegenerationDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderNewToolRegenerationDialog(c, false, data)
}

func PutToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	tr, merr := db.GetToolRegeneration(shared.EntityID(id))
	if merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to load tool regeneration with ID %d: %v", id, merr))
		return reRenderEditToolRegenerationDialog(c, id, true, ToolRegenerationFormData{}, ierr)
	}

	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderEditToolRegenerationDialog(c, id, true, data, ierrs...)
	}

	tr.Start = data.Start
	tr.Stop = data.Stop

	if verr := tr.Validate(); verr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("validation error: %v", verr))
		return reRenderEditToolRegenerationDialog(c, id, true, data, ierr)
	}

	if merr := db.UpdateToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to update tool regeneration: %v", merr))
		return reRenderEditToolRegenerationDialog(c, id, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderEditToolRegenerationDialog(c, id, false, data)
}

func parseToolRegenerationForm(c echo.Context) (data ToolRegenerationFormData, ierrs []*errors.InputError) {
	id, err := utils.SanitizeInt64(c.FormValue("tool_id"))
	if err != nil {
		ierr := errors.NewInputError("tool_id", fmt.Sprintf("invalid tool ID: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.ToolID = shared.EntityID(id)

	if data.ToolID == 0 {
		ierr := errors.NewInputError("tool_id", "tool ID is required")
		ierrs = append(ierrs, ierr)
	}

	startStr := c.FormValue("start")
	if startStr != "" {
		startTime, err := utils.ParseDate(startStr)
		if err != nil {
			ierr := errors.NewInputError("start", fmt.Sprintf("invalid start date: %v", err))
			ierrs = append(ierrs, ierr)
		} else {
			data.Start = shared.NewUnixMilli(startTime)
		}
	}

	stopStr := c.FormValue("stop")
	if stopStr != "" {
		stopTime, err := utils.ParseDate(stopStr)
		if err != nil {
			ierr := errors.NewInputError("stop", fmt.Sprintf("invalid stop date: %v", err))
			ierrs = append(ierrs, ierr)
		} else {
			data.Stop = shared.NewUnixMilli(stopTime)
		}
	}

	return
}

func reRenderNewToolRegenerationDialog(c echo.Context, open bool, formData ToolRegenerationFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := NewToolRegenerationDialog(ToolRegenerationDialogProps{
		ToolRegenerationFormData: formData,
		Open:                     open,
		OOB:                      true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}

func reRenderEditToolRegenerationDialog(c echo.Context, regenID shared.EntityID, open bool, formData ToolRegenerationFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := EditToolRegenerationDialog(regenID, ToolRegenerationDialogProps{
		ToolRegenerationFormData: formData,
		Open:                     open,
		OOB:                      true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}

	if c.QueryParam("tool_id") == "" {
		return echo.NewHTTPError(400, "missing tool or regeneration ID")
	}

	id, merr := utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}

	tr, merr := db.GetToolRegeneration(shared.EntityID(id))
	if merr != nil {
		ierr := errors.NewInputError("", "failed to load tool regeneration with ID "+shared.EntityID(id).String()+": "+merr.Error())
		return reRenderEditToolRegenerationDialog(c, shared.EntityID(id), true, ToolRegenerationFormData{}, ierr)
	}

	return reRenderEditToolRegenerationDialog(c, shared.EntityID(id), true, ToolRegenerationFormData{
		Start: tr.Start,
		Stop:  tr.Stop,
	})
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderNewToolRegenerationDialog(c, true, data, ierrs...)
	}

	tr := &shared.ToolRegeneration{
		Start: data.Start,
		Stop:  data.Stop,
	}
	if merr := db.AddToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", "failed to create tool regeneration: "+merr.Error())
		return reRenderNewToolRegenerationDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderNewToolRegenerationDialog(c, false, data)
}

	if c.QueryParam("tool_id") == "" {

	if c.QueryParam("tool_id") == "" {
		return echo.NewHTTPError(400, "missing tool or regeneration ID")
	}

	id, merr := utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}

	t := NewToolRegenerationDialog(ToolRegenerationDialogProps{
		Open: true,
		OOB:  true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	return nil
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	if id, _ := utils.GetQueryInt64(c, "id"); id != 0 {
		return updateToolRegeneration(c, shared.EntityID(id))
	}

	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderNewToolRegenerationDialog(c, true, data, ierrs...)
	}

	tr := &shared.ToolRegeneration{
		Start: data.Start,
		Stop:  data.Stop,
	}
	if merr := db.AddToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", "failed to create tool regeneration: "+merr.Error())
		return reRenderNewToolRegenerationDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderNewToolRegenerationDialog(c, false, data)
}

func updateToolRegeneration(c echo.Context, toolRegenID shared.EntityID) *echo.HTTPError {
	tr, merr := db.GetToolRegeneration(toolRegenID)
	if merr != nil {
		ierr := errors.NewInputError("", "failed to load tool regeneration with ID "+toolRegenID.String()+": "+merr.Error())
		return reRenderEditToolRegenerationDialog(c, toolRegenID, true, ToolRegenerationFormData{}, ierr)
	}

	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderEditToolRegenerationDialog(c, toolRegenID, true, data, ierrs...)
	}
	tr.Start = data.Start
	tr.Stop = data.Stop

	if merr := db.UpdateToolRegeneration(tr); merr != nil {
		ierr := errors.NewInputError("", "failed to update tool regeneration: "+merr.Error())
		return reRenderEditToolRegenerationDialog(c, toolRegenID, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-tool-regenerations")

	return reRenderEditToolRegenerationDialog(c, toolRegenID, false, data)
}

func parseToolRegenerationForm(c echo.Context) (data ToolRegenerationFormData, ierrs []*errors.InputError) {
	startTime, err := time.Parse("2006-01-02", c.FormValue("start"))
	if err != nil {
		ierr := errors.NewInputError("start", "invalid start time: "+err.Error())
		ierrs = append(ierrs, ierr)
	}
	data.Start = shared.NewUnixMilli(startTime)

	stopTime, err := time.Parse("2006-01-02", c.FormValue("stop"))
	if err != nil {
		ierr := errors.NewInputError("stop", "invalid stop time: "+err.Error())
		ierrs = append(ierrs, ierr)
	}
	data.Stop = shared.NewUnixMilli(stopTime)

	return
}

func reRenderNewToolRegenerationDialog(c echo.Context, open bool, formData ToolRegenerationFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := NewToolRegenerationDialog(ToolRegenerationDialogProps{
		ToolRegenerationFormData: formData,
		Open:                     open,
		OOB:                      true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolRegenerationDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(400, "invalid input")
	}

	return nil
}

func reRenderEditToolRegenerationDialog(c echo.Context, toolRegenID shared.EntityID, open bool, formData ToolRegenerationFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := EditToolRegenerationDialog(toolRegenID, ToolRegenerationDialogProps{
		ToolRegenerationFormData: formData,
		Open:                     open,
		OOB:                      true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(400, "invalid input")
	}

	return nil
}
