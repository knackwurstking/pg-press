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
	tool, verr := getCassetteDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	log.Debug("Creating new cassette: %#v", tool.String())

	merr := db.AddTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutCassette handles updating an existing cassette
func PutCassette(c echo.Context) *echo.HTTPError {
	tool, verr := getCassetteDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool.ID = shared.EntityID(id)

	log.Debug("Updating cassette: %v", tool.String())

	merr = db.UpdateTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	urlb.SetHXRedirect(c, urlb.UrlTool(tool.ID, 0, 0).Page)

	return nil
}

func getCassetteDialogForm(c echo.Context) (*shared.Tool, *errors.ValidationError) {
	var (
		vWidth        = c.FormValue("width")
		vHeight       = c.FormValue("height")
		vType         = strings.Trim(c.FormValue("type"), " ")
		vCode         = strings.Trim(c.FormValue("code"), " ")
		vMinThickness = c.FormValue("min_thickness")
		vMaxThickness = c.FormValue("max_thickness")
	)

	log.Debug("Cassette dialog form values: width=%s, height=%s, type=%s, code=%s, min_thickness=%s, max_thickness=%s",
		vWidth, vHeight, vType, vCode, vMinThickness, vMaxThickness)

	// Convert vWidth and vHeight to integers
	width, err := strconv.Atoi(vWidth)
	if err != nil {
		return nil, errors.NewValidationError("invalid width: %s", vWidth)
	}
	height, err := strconv.Atoi(vHeight)
	if err != nil {
		return nil, errors.NewValidationError("invalid height: %s", vHeight)
	}

	// Convert thickness values to floats
	minThickness, err := strconv.ParseFloat(vMinThickness, 32)
	if err != nil {
		return nil, errors.NewValidationError("invalid min thickness: %s", vMinThickness)
	}
	maxThickness, err := strconv.ParseFloat(vMaxThickness, 32)
	if err != nil {
		return nil, errors.NewValidationError("invalid max thickness: %s", vMaxThickness)
	}

	// Type and Code have to be set
	if vType == "" {
		return nil, errors.NewValidationError("type is required")
	}
	if vCode == "" {
		return nil, errors.NewValidationError("code is required")
	}

	cassette := &shared.Tool{
		Width:        width,
		Height:       height,
		Position:     shared.SlotUpperCassette,
		Type:         vType,
		Code:         vCode,
		CyclesOffset: 0, // TODO: Maybe update the dialog to allow changing this?
		MinThickness: float32(minThickness),
		MaxThickness: float32(maxThickness),
	}

	if verr := cassette.Validate(); verr != nil {
		return cassette, verr
	}

	return cassette, nil
}
