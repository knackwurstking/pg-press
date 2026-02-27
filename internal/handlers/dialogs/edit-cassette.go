package dialogs

import (
	"fmt"
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
		log.Debug("Rendering edit cassette dialog: %#v", tool.String())
		t := EditCassetteDialog(EditCassetteDialogProps{
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

	log.Debug("Rendering new cassette dialog...")
	t := NewCassetteDialog(NewCassetteDialogProps{
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
		updateCassette(c, shared.EntityID(id))
	}

	formData, ierr := parseCassetteForm(c)
	if ierr != nil {
		t := NewCassetteDialog(NewCassetteDialogProps{
			CassetteFormData: formData,
			Open:             true,
			OOB:              true,
			Error:            ierr,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "NewCassetteDialog")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input: %s", ierr.Error())
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

	log.Debug("Creating new cassette: %#v", tool.String())

	if merr := db.AddTool(tool); merr != nil {
		ierr = errors.NewInputError("form", fmt.Sprintf("Failed to create cassette: %s", merr.Error()))
		t := NewCassetteDialog(NewCassetteDialogProps{
			CassetteFormData: formData,
			Open:             true,
			OOB:              true,
			Error:            ierr,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "NewCassetteDialog")
		}
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	t := NewCassetteDialog(NewCassetteDialogProps{
		Open:  false,
		OOB:   true,
		Error: ierr,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}
	return nil
}

func updateCassette(c echo.Context, toolID shared.EntityID) *echo.HTTPError {
	formData, ierr := parseCassetteForm(c)
	if ierr != nil {
		t := EditCassetteDialog(EditCassetteDialogProps{
			CassetteFormData: formData,
			Open:             true,
			OOB:              true,
			Error:            ierr,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input: %s", ierr.Error())
	}

	tool, merr := db.GetTool(shared.EntityID(toolID))
	if merr != nil {
		ierr := errors.NewInputError("id", fmt.Sprintf("Cassette with ID %d not found", toolID))
		t := EditCassetteDialog(EditCassetteDialogProps{
			CassetteFormData: formData,
			Open:             true,
			OOB:              true,
			Error:            ierr,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return merr.Echo()
	}
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Width = formData.Width
	tool.Height = formData.Height
	tool.MinThickness = formData.MinThickness
	tool.MaxThickness = formData.MaxThickness

	log.Debug("Updating cassette: %#v", tool.String())

	if merr = db.UpdateTool(tool); merr != nil {
		ierr = errors.NewInputError("form", fmt.Sprintf("Failed to update cassette: %s", merr.Error()))
		t := EditCassetteDialog(EditCassetteDialogProps{
			CassetteFormData: formData,
			Open:             true,
			OOB:              true,
			Error:            ierr,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return merr.Echo()
	}

	// Set HX headers
	utils.SetHXRedirect(c, urlb.Tool(tool.ID))

	t := EditCassetteDialog(EditCassetteDialogProps{
		Open:  false,
		OOB:   true,
		Error: ierr,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "EditCassetteDialog")
	}
	return nil
}

func parseCassetteForm(c echo.Context) (data CassetteFormData, ierr *errors.InputError) {
	// Sanitize inputs by trimming whitespace
	data.Type = utils.SanitizeText(c.FormValue("type"))
	data.Code = utils.SanitizeText(c.FormValue("code"))

	// Convert vWidth and vHeight to integers with sanitization
	var err error
	data.Width, err = utils.SanitizeInt(c.FormValue("width"))
	if err != nil {
		ierr = errors.NewInputError("width", "Invalid width: must be an integer")
		return
	}

	data.Height, err = utils.SanitizeInt(c.FormValue("height"))
	if err != nil {
		ierr = errors.NewInputError("height", "Invalid height: must be an integer")
		return
	}

	// Convert thickness values to floats with sanitization, min thickness can be zero
	minThickness, _ := utils.SanitizeFloat(c.FormValue("min-thickness"))
	data.MinThickness = float32(minThickness)

	maxThickness, err := utils.SanitizeFloat(c.FormValue("max-thickness"))
	if err != nil {
		ierr = errors.NewInputError("max-thickness", "Invalid max thickness: must be a valid number")
		return
	} else if maxThickness <= 0 {
		ierr = errors.NewInputError("max-thickness", "Max thickness must be greater than zero")
		return
	}
	data.MaxThickness = float32(maxThickness)

	log.Debug("Cassette dialog form values: %#v", data)

	return
}
