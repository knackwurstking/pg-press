package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handler/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/models"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetToolDialog(c echo.Context) *echo.HTTPError {
	var tool *shared.Tool
	id, _ := shared.ParseQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.MasterError
		tool, merr = h.DB.Tool.Tool.GetByID(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
	}

	var t templ.Component
	var tName string
	if tool != nil {
		t = templates.EditToolDialog(tool)
		tName = "EditToolDialog"
		h.Logger.Debug("Rendering edit tool dialog: %#v", tool.String())
	} else {
		t = templates.NewToolDialog()
		tName = "NewToolDialog"
		h.Logger.Debug("Rendering new tool dialog...")
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func (h *Handler) PostTool(c echo.Context) *echo.HTTPError {
	tool, verr := GetToolDialogForm(c)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	h.Logger.Debug("Creating new tool: %#v", tool.String())

	merr := h.DB.Tool.Tool.Create(tool)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutTool handles updating an existing tool
func (h *Handler) PutTool(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	// TODO: Continue here...
	tool, merr := GetToolDialogForm(c)
	if merr != nil {
		return merr.Echo()
	}
	tool.ID = toolID // Just to be sure

	h.Logger.Debug("Updating tool: %#v", tool.String())

	merr = h.DB.Tool.Tool.Update(tool)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	urlb.SetHXRedirect(c, urlb.UrlTool(tool.ID, 0, 0).Page)

	// TODO: Needs to be removed, just do a redirect to the same page to keep it simple
	//utils.SetHXAfterSettle(c, map[string]any{
	//	"toolUpdated": map[string]string{
	//		"pageTitle": fmt.Sprintf("PG Presse | %s %s",
	//			tool.String(), tool.Position.GermanString()),
	//		"appBarTitle": fmt.Sprintf("%s %s", tool.String(),
	//			tool.Position.GermanString()),
	//	},
	//})

	return nil
}

func GetToolDialogForm(c echo.Context) (*shared.Tool, *errors.ValidationError) {
	var (
		vPosition = c.FormValue("position")
		vWidth    = c.FormValue("width")
		vHeight   = c.FormValue("height")
		vType     = strings.Trim(c.FormValue("type"), " ")
		vCode     = strings.Trim(c.FormValue("code"), " ")
	)

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
