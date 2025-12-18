package dialogs

import (
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func GetToolDialog(c echo.Context) *echo.HTTPError {
	var tool *shared.Tool
	id, _ := shared.ParseQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.MasterError
		tool, merr = DB.Tool.Tool.GetByID(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
	}

	var t templ.Component
	var tName string
	if tool != nil {
		t = EditToolDialog(tool)
		tName = "EditToolDialog"
		Log.Debug("Rendering edit tool dialog: %#v", tool.String())
	} else {
		t = NewToolDialog()
		tName = "NewToolDialog"
		Log.Debug("Rendering new tool dialog...")
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func PostTool(c echo.Context) *echo.HTTPError {
	tool, verr := getToolDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	Log.Debug("Creating new tool: %#v", tool.String())

	merr := DB.Tool.Tool.Create(tool)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutTool handles updating an existing tool
func PutTool(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	tool, verr := getToolDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}
	tool.ID = toolID // Just to be sure

	Log.Debug("Updating tool: %#v", tool.String())

	merr = DB.Tool.Tool.Update(tool)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	urlb.SetHXRedirect(c, urlb.UrlTool(tool.ID, 0, 0).Page)

	return nil
}

func getToolDialogForm(c echo.Context) (*shared.Tool, *errors.ValidationError) {
	var (
		vPosition = c.FormValue("position")
		vWidth    = c.FormValue("width")
		vHeight   = c.FormValue("height")
		vType     = strings.Trim(c.FormValue("type"), " ")
		vCode     = strings.Trim(c.FormValue("code"), " ")
	)

	Log.Debug("Tool dialog form values: position=%s, width=%s, height=%s, type=%s, code=%s",
		vPosition, vWidth, vHeight, vType, vCode)

	// Need to convert the vPosition to an integer
	position, err := strconv.Atoi(vPosition)
	if err != nil {
		return nil, errors.NewValidationError("invalid position: %s", vPosition)
	}

	// Check and set position
	switch shared.Slot(position) {
	case shared.SlotUpper, shared.SlotLower:
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

	// Type and Code have to be set
	if vType == "" {
		return nil, errors.NewValidationError("type is required")
	}
	if vCode == "" {
		return nil, errors.NewValidationError("code is required")
	}

	tool := &shared.Tool{
		BaseTool: shared.BaseTool{
			Width:        width,
			Height:       height,
			Position:     shared.Slot(position),
			Type:         vType,
			Code:         vCode,
			CyclesOffset: 0, // TODO: Maybe update the dialog to allow changing this?
		},
	}

	if verr := tool.Validate(); verr != nil {
		return tool, verr
	}

	return tool, nil
}
