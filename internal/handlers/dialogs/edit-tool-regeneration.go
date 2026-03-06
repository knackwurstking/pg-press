package dialogs

import (
	"fmt"
	"net/http"
	"time"

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
	if id, _ := utils.GetQueryInt64(c, "id"); id != 0 {
		return updateToolRegeneration(c, shared.EntityID(id))
	}

	data, ierrs := parseToolRegenerationForm(c)
	if len(ierrs) > 0 {
		return reRenderNewToolRegenerationDialog(c, true, data, ierrs...)
	}

	tr := &shared.ToolRegeneration{
		ToolID: data.ToolID,
		Start:  data.Start,
		Stop:   data.Stop,
	}
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

func parseToolRegenerationForm(c echo.Context) (data ToolRegenerationFormData, ierrs []*errors.InputError) {
	toolID, err := utils.SanitizeInt64(c.FormValue("tool_id"))
	if err != nil {
		ierr := errors.NewInputError("tool_id", fmt.Sprintf("invalid tool ID: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.ToolID = shared.EntityID(toolID)

	startTime, err := time.Parse("2006-01-02", c.FormValue("start"))
	if err != nil {
		ierr := errors.NewInputError("start", fmt.Sprintf("invalid start date: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.Start = shared.NewUnixMilli(startTime)

	stopTime, err := time.Parse("2006-01-02", c.FormValue("stop"))
	if err != nil {
		ierr := errors.NewInputError("stop", fmt.Sprintf("invalid stop date: %v", err))
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
		Error:                    ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
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
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}
