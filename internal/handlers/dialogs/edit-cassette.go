package dialogs

import (
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetCassetteDialog(c echo.Context) *echo.HTTPError {
	var cassette *shared.Cassette
	id, _ := shared.ParseQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.MasterError
		cassette, merr = h.DB.Tool.Cassette.GetByID(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
	}

	var t templ.Component
	var tName string
	if cassette != nil {
		t = templates.EditCassetteDialog(cassette)
		tName = "EditCassetteDialog"
		h.Log.Debug("Rendering edit cassette dialog: %#v", cassette.String())
	} else {
		t = templates.NewCassetteDialog()
		tName = "NewCassetteDialog"
		h.Log.Debug("Rendering new cassette dialog...")
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func (h *Handler) PostCassette(c echo.Context) *echo.HTTPError {
	cassette, verr := h.getCassetteDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	h.Log.Debug("Creating new cassette: %#v", cassette.String())

	merr := h.DB.Tool.Cassette.Create(cassette)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutCassette handles updating an existing cassette
func (h *Handler) PutCassette(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	cassetteID := shared.EntityID(id)

	cassette, verr := h.getCassetteDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}
	cassette.ID = cassetteID

	h.Log.Debug("Updating cassette: %#v", cassette.String())

	merr = h.DB.Tool.Cassette.Update(cassette)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	urlb.SetHXRedirect(c, urlb.UrlTool(cassette.ID, 0, 0, 0).Page)

	return nil
}

func (h *Handler) getCassetteDialogForm(c echo.Context) (*shared.Cassette, *errors.ValidationError) {
	var (
		vPosition     = c.FormValue("position")
		vWidth        = c.FormValue("width")
		vHeight       = c.FormValue("height")
		vType         = strings.Trim(c.FormValue("type"), " ")
		vCode         = strings.Trim(c.FormValue("code"), " ")
		vMinThickness = c.FormValue("min_thickness")
		vMaxThickness = c.FormValue("max_thickness")
	)

	h.Log.Debug("Cassette dialog form values: position=%s, width=%s, height=%s, type=%s, code=%s, min_thickness=%s, max_thickness=%s",
		vPosition, vWidth, vHeight, vType, vCode, vMinThickness, vMaxThickness)

	// Need to convert the vPosition to an integer
	position, err := strconv.Atoi(vPosition)
	if err != nil {
		return nil, errors.NewValidationError("invalid position: %s", vPosition)
	}

	// Check and set position
	switch shared.Slot(position) {
	case shared.SlotUpper, shared.SlotLower, shared.SlotUpperCassette:
	default:
		return nil, errors.NewValidationError("invalid position: %s", vPosition)
	}

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

	cassette := &shared.Cassette{
		BaseTool: shared.BaseTool{
			Width:        width,
			Height:       height,
			Position:     shared.Slot(position),
			Type:         vType,
			Code:         vCode,
			CyclesOffset: 0, // TODO: Maybe update the dialog to allow changing this?
		},
		MinThickness: float32(minThickness),
		MaxThickness: float32(maxThickness),
	}

	if verr := cassette.Validate(); verr != nil {
		return cassette, verr
	}

	return cassette, nil
}
