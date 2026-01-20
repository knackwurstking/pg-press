package dialogs

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
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
		t := templates.EditCassetteDialog(tool)
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditCassetteDialog")
		}
		return nil
	}

	log.Debug("Rendering new cassette dialog...")
	t := templates.NewCassetteDialog()
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewCassetteDialog")
	}

	return nil
}

func PostCassette(c echo.Context) *echo.HTTPError {
	tool, verr := parseCassetteForm(c, nil)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	log.Debug("Creating new cassette: %v", tool.String())

	merr := db.AddTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutCassette handles updating an existing cassette
func PutCassette(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	tool, verr := parseCassetteForm(c, tool)
	if verr != nil {
		return verr.HTTPError().Echo()
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

func parseCassetteForm(c echo.Context, tool *shared.Tool) (*shared.Tool, *errors.ValidationError) {
	if tool == nil {
		tool = &shared.Tool{
			Position: shared.SlotUpperCassette,
		}
	}

	// Sanitize inputs by trimming whitespace
	tool.Type = strings.TrimSpace(c.FormValue("type"))
	tool.Code = strings.TrimSpace(c.FormValue("code"))

	// Convert vWidth and vHeight to integers with sanitization
	var err error
	widthStr := strings.TrimSpace(c.FormValue("width"))
	tool.Width, err = strconv.Atoi(widthStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid width: %s", widthStr)
	}

	heightStr := strings.TrimSpace(c.FormValue("height"))
	tool.Height, err = strconv.Atoi(heightStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid height: %s", heightStr)
	}

	// Convert thickness values to floats with sanitization
	minThicknessStr := strings.TrimSpace(c.FormValue("min_thickness"))
	minThickness, err := strconv.ParseFloat(minThicknessStr, 32)
	if err != nil {
		return nil, errors.NewValidationError("invalid min thickness: %s", minThicknessStr)
	}
	tool.MinThickness = float32(minThickness)

	maxThicknessStr := strings.TrimSpace(c.FormValue("max_thickness"))
	maxThickness, err := strconv.ParseFloat(maxThicknessStr, 32)
	if err != nil {
		return nil, errors.NewValidationError("invalid max thickness: %s", maxThicknessStr)
	}
	tool.MaxThickness = float32(maxThickness)

	log.Debug("Cassette dialog form values: tool=%v", tool)

	if verr := tool.Validate(); verr != nil {
		return tool, verr
	}

	return tool, nil
}
