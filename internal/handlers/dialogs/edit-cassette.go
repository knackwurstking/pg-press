package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetCassetteDialog(c echo.Context) *echo.HTTPError {
	var tool *shared.Tool
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.HTTPError
		tool, merr = db.GetTool(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
		if !tool.IsCassette() {
			return echo.NewHTTPError(http.StatusBadRequest, "tool with ID %d is not a cassette", id)
		}
	}

	if tool != nil {
		slog.Debug("Rendering edit cassette dialog", "tool_string", tool.String())
		t := EditCassetteDialog(tool.ID, CassetteDialogProps{
			CassetteFormData: CassetteFormData{
				Type:         tool.Type,
				Code:         tool.Code,
				Width:        tool.Width,
				Height:       tool.Height,
				MinThickness: tool.MinThickness,
				MaxThickness: tool.MaxThickness,
			},
			OOB:  true,
			Open: true,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return nil
	}

	slog.Debug("Rendering new cassette dialog...")
	t := NewCassetteDialog(CassetteDialogProps{
		OOB:  true,
		Open: true,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}

	return nil
}

func PostCassette(c echo.Context) *echo.HTTPError {
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		return updateCassette(c, shared.EntityID(id))
	}

	formData, ierrs := parseCassetteForm(c)
	if len(ierrs) > 0 {
		return reRenderNewCassetteDialog(c, true, formData, ierrs...)
	}

	tool := &shared.Tool{
		Position:     shared.SlotUpperCassette,
		Type:         formData.Type,
		Code:         formData.Code,
		Width:        formData.Width,
		Height:       formData.Height,
		MinThickness: formData.MinThickness,
		MaxThickness: formData.MaxThickness,
	}

	slog.Debug("Creating new cassette", "tool_string", tool.String())

	if merr := db.AddTool(tool); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("Failed to create cassette: %s", merr.Error()))
		return reRenderNewCassetteDialog(c, true, formData, ierr)
	}

	utils.SetHXTrigger(c, "tool-tab-content")

	return reRenderNewCassetteDialog(c, false, formData, nil)
}

func updateCassette(c echo.Context, toolID shared.EntityID) *echo.HTTPError {
	formData, ierrs := parseCassetteForm(c)
	if len(ierrs) > 0 {
		return reRenderEditCassetteDialog(c, toolID, true, formData, ierrs...)
	}

	tool, merr := db.GetTool(shared.EntityID(toolID))
	if merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("Cassette with ID %d not found", toolID))
		return reRenderEditCassetteDialog(c, toolID, true, formData, ierr)
	}
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Width = formData.Width
	tool.Height = formData.Height
	tool.MinThickness = formData.MinThickness
	tool.MaxThickness = formData.MaxThickness

	slog.Debug("Updating cassette", "tool", tool)

	if merr = db.UpdateTool(tool); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("Failed to update cassette: %s", merr.Error()))
		return reRenderEditCassetteDialog(c, toolID, true, formData, ierr)
	}

	// Set HX headers
	utils.SetHXRedirect(c, urlb.Tool(tool.ID))

	return reRenderEditCassetteDialog(c, toolID, false, formData)
}

func parseCassetteForm(c echo.Context) (data CassetteFormData, ierrs []*errors.InputError) {
	// Sanitize inputs by trimming whitespace
	data.Type = utils.SanitizeText(c.FormValue("type"))
	data.Code = utils.SanitizeText(c.FormValue("code"))

	// Convert vWidth and vHeight to integers with sanitization
	var err error
	data.Width, err = utils.SanitizeInt(c.FormValue("width"))
	if err != nil {
		ierr := errors.NewInputError("width", "Invalid width: must be an integer")
		ierrs = append(ierrs, ierr)
	}

	data.Height, err = utils.SanitizeInt(c.FormValue("height"))
	if err != nil {
		ierr := errors.NewInputError("height", "Invalid height: must be an integer")
		ierrs = append(ierrs, ierr)
	}

	// Convert thickness values to floats with sanitization, min thickness can be zero
	minThickness, _ := utils.SanitizeFloat(c.FormValue("min-thickness"))
	data.MinThickness = float32(minThickness)

	maxThickness, err := utils.SanitizeFloat(c.FormValue("max-thickness"))
	if err != nil {
		ierr := errors.NewInputError("max-thickness", "Invalid max thickness: must be a valid number")
		ierrs = append(ierrs, ierr)
	} else if maxThickness <= 0 {
		ierr := errors.NewInputError("max-thickness", "Max thickness must be greater than zero")
		ierrs = append(ierrs, ierr)
	}
	data.MaxThickness = float32(maxThickness)

	slog.Debug("Cassette dialog form values", "data", data)

	return
}

func reRenderNewCassetteDialog(c echo.Context, open bool, data CassetteFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := NewCassetteDialog(CassetteDialogProps{
		CassetteFormData: data,
		Open:             open,
		OOB:              true,
		Error:            ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input"))
	}
	return nil
}

func reRenderEditCassetteDialog(c echo.Context, toolID shared.EntityID, open bool, data CassetteFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := EditCassetteDialog(toolID, CassetteDialogProps{
		CassetteFormData: data,
		Open:             open,
		OOB:              true,
		Error:            ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "EditCassetteDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input"))
	}
	return nil
}
