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
		t := EditCassetteDialog(true, tool, nil)
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return nil
	}

	log.Debug("Rendering new cassette dialog...")
	t := NewCassetteDialog(true, true, nil)
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}

	return nil
}

func PostCassette(c echo.Context) *echo.HTTPError {
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		updateCassette(c)
	}

	tool, ierr := parseCassetteForm(c, nil)
	if ierr != nil {
		t := NewCassetteDialog(true, true, ierr)
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "NewCassetteDialog")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input: %s", ierr.Error())
	}

	log.Debug("Creating new cassette: %v", tool.String())

	merr := db.AddTool(tool)
	if merr != nil {
		ierr = errors.NewInputError("form", fmt.Sprintf("Failed to create cassette: %s", merr.Error()))
		t := NewCassetteDialog(true, true, ierr)
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "NewCassetteDialog")
		}
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	t := NewCassetteDialog(false, true, ierr) // TODO: Close...
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}
	return nil
}

func updateCassette(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	tool, ierr := parseCassetteForm(c, tool)
	if ierr != nil {
		// TODO: Re-render dialog with edited content and error message (OOB)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input: %s", ierr.Error())
	}

	log.Debug("Updating cassette: %v", tool.String())

	merr = db.UpdateTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	utils.SetHXRedirect(c, urlb.Tool(tool.ID))

	return nil
}

func parseCassetteForm(c echo.Context, tool *shared.Tool) (*shared.Tool, *errors.InputError) {
	if tool == nil {
		tool = &shared.Tool{
			Position: shared.SlotUpperCassette,
		}
	}

	// Sanitize inputs by trimming whitespace
	tool.Type = utils.SanitizeText(c.FormValue("type"))
	tool.Code = utils.SanitizeText(c.FormValue("code"))

	// Convert vWidth and vHeight to integers with sanitization
	var err error
	tool.Width, err = utils.SanitizeInt(c.FormValue("width"))
	if err != nil {
		return nil, errors.NewInputError("width", "Invalid width: must be an integer")
	}

	tool.Height, err = utils.SanitizeInt(c.FormValue("height"))
	if err != nil {
		return nil, errors.NewInputError("height", "Invalid height: must be an integer")
	}

	// Convert thickness values to floats with sanitization, min thickness can be zero
	minThickness, _ := utils.SanitizeFloat(c.FormValue("min-thickness"))
	tool.MinThickness = float32(minThickness)

	maxThickness, err := utils.SanitizeFloat(c.FormValue("max-thickness"))
	if err != nil {
		return nil, errors.NewInputError("max-thickness", "Invalid max thickness: must be a valid number")
	} else if maxThickness <= 0 {
		return nil, errors.NewInputError("max-thickness", "Max thickness must be greater than zero")
	}
	tool.MaxThickness = float32(maxThickness)

	log.Debug("Cassette dialog form values: tool=%v", tool)

	if verr := tool.Validate(); verr != nil {
		return tool, errors.NewInputError("form", verr.Error())
	}

	return tool, nil
}
