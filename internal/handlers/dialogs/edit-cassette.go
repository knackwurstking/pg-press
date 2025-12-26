package dialogs

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetCassetteDialog(c echo.Context) *echo.HTTPError {
	var tool *shared.Tool
	id, _ := shared.ParseQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.MasterError
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
		t := EditCassetteDialog(tool)
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return nil
	}

	log.Debug("Rendering new cassette dialog...")
	t := NewCassetteDialog()
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}

	return nil
}

func PostCassette(c echo.Context) *echo.HTTPError {
	tool, verr := parseCassetteForm(c, nil)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	log.Debug("Creating new cassette: %v", tool.String())

	merr := db.AddTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutCassette handles updating an existing cassette
func PutCassette(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	tool, verr := parseCassetteForm(c, tool)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	log.Debug("Updating cassette: %v", tool.String())

	merr = db.UpdateTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	urlb.SetHXRedirect(c, urlb.UrlTool(tool.ID, 0, 0).Page)

	return nil
}

func parseCassetteForm(c echo.Context, tool *shared.Tool) (*shared.Tool, *errors.ValidationError) {
	if tool == nil {
		tool = &shared.Tool{
			Position: shared.SlotUpperCassette,
		}
	}
	tool.Type = strings.Trim(c.FormValue("type"), " ")
	tool.Code = strings.Trim(c.FormValue("code"), " ")

	// Convert vWidth and vHeight to integers
	var err error
	tool.Width, err = strconv.Atoi(c.FormValue("width"))
	if err != nil {
		return nil, errors.NewValidationError("invalid width: %s", c.FormValue("width"))
	}
	tool.Height, err = strconv.Atoi(c.FormValue("height"))
	if err != nil {
		return nil, errors.NewValidationError("invalid height: %s", c.FormValue("height"))
	}

	// Convert thickness values to floats
	minThickness, err := strconv.ParseFloat(c.FormValue("min_thickness"), 32)
	if err != nil {
		return nil, errors.NewValidationError("invalid min thickness: %s", c.FormValue("min_thickness"))
	}
	tool.MinThickness = float32(minThickness)

	maxThickness, err := strconv.ParseFloat(c.FormValue("max_thickness"), 32)
	if err != nil {
		return nil, errors.NewValidationError("invalid max thickness: %s", c.FormValue("max_thickness"))
	}
	tool.MaxThickness = float32(maxThickness)

	log.Debug("Cassette dialog form values: tool=%v", tool)

	if verr := tool.Validate(); verr != nil {
		return tool, verr
	}

	return tool, nil
}
