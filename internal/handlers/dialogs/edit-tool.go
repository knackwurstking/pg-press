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

func GetToolDialog(c echo.Context) *echo.HTTPError {
	var tool *shared.Tool
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.HTTPError
		tool, merr = db.GetTool(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
	}

	if tool != nil {
		log.Debug("Rendering edit tool dialog: %#v", tool)
		t := EditToolDialog(tool.ID, ToolDialogProps{
			ToolFormData: ToolFormData{
				Type:     tool.Type,
				Code:     tool.Code,
				Position: tool.Position,
				Width:    tool.Width,
				Height:   tool.Height,
			},
			OOB:  true,
			Open: true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolDialog")
		}
		return nil
	}

	log.Debug("Rendering new tool dialog...")
	t := NewToolDialog(ToolDialogProps{
		Open: true,
		OOB:  true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolDialog")
	}

	return nil
}

func PostTool(c echo.Context) *echo.HTTPError {
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		return updateTool(c, shared.EntityID(id))
	}

	formData, ierrs := parseToolForm(c)
	if len(ierrs) > 0 {
		return reRenderNewToolDialog(c, true, formData, ierrs...)
	}
	log.Debug("Creating new tool: %#v", formData)

	tool := &shared.Tool{
		Type:     formData.Type,
		Code:     formData.Code,
		Position: formData.Position,
		Width:    formData.Width,
		Height:   formData.Height,
	}
	if merr := db.AddTool(tool); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("Failed to create tool: %s", merr.Error()))
		return reRenderNewToolDialog(c, true, formData, ierr)
	}

	utils.SetHXTrigger(c, "tool-tab-content")

	return reRenderNewToolDialog(c, false, formData, nil)
}

func updateTool(c echo.Context, toolID shared.EntityID) *echo.HTTPError {
	formData, ierrs := parseToolForm(c)
	if len(ierrs) > 0 {
		return reRenderEditToolDialog(c, toolID, true, formData, ierrs...)
	}

	tool, merr := db.GetTool(toolID)
	if merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("Failed to load tool: %s", merr.Error()))
		return reRenderEditToolDialog(c, toolID, true, formData, ierr)
	}
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Position = formData.Position
	tool.Width = formData.Width
	tool.Height = formData.Height

	log.Debug("Updating tool: %#v", tool)

	if merr = db.UpdateTool(tool); merr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("Failed to update tool: %s", merr.Error()))
		return reRenderEditToolDialog(c, toolID, true, formData, ierr)
	}

	// Set HX headers
	utils.SetHXRedirect(c, urlb.Tool(tool.ID))

	// Close dialog
	return reRenderEditToolDialog(c, toolID, false, formData, nil)
}

func parseToolForm(c echo.Context) (data ToolFormData, ierrs []*errors.InputError) {
	// Sanitize inputs by trimming whitespace
	data.Type = utils.SanitizeText(c.FormValue("type"))
	data.Code = utils.SanitizeText(c.FormValue("code"))

	// Need to convert the vPosition to an integer
	position, err := utils.SanitizeInt(c.FormValue("position"))
	if err != nil {
		ierr := errors.NewInputError("position", fmt.Sprintf("Invalid position: %s", c.FormValue("position")))
		ierrs = append(ierrs, ierr)
	}
	switch v := shared.Slot(position); v {
	case shared.SlotUpper, shared.SlotLower:
		data.Position = shared.Slot(position)
	default:
		ierr := errors.NewInputError("position", fmt.Sprintf("Invalid position: %s", c.FormValue("position")))
		ierrs = append(ierrs, ierr)
	}

	// Convert width and height to integers with sanitization
	data.Width, err = utils.SanitizeInt(c.FormValue("width"))
	if err != nil {
		ierr := errors.NewInputError("width", fmt.Sprintf("Invalid width: %s", c.FormValue("width")))
		ierrs = append(ierrs, ierr)
	}

	data.Height, err = utils.SanitizeInt(c.FormValue("height"))
	if err != nil {
		ierr := errors.NewInputError("height", fmt.Sprintf("Invalid height: %s", c.FormValue("height")))
		ierrs = append(ierrs, ierr)
	}

	log.Debug("Tool dialog form values: %#v", data)

	return
}

func reRenderNewToolDialog(c echo.Context, open bool, data ToolFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := NewToolDialog(ToolDialogProps{
		ToolFormData: data,
		Open:         open,
		OOB:          true,
		Error:        ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewToolDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}
	return nil
}

func reRenderEditToolDialog(c echo.Context, toolID shared.EntityID, open bool, data ToolFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := EditToolDialog(toolID, ToolDialogProps{
		ToolFormData: data,
		Open:         open,
		OOB:          true,
		Error:        ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "EditToolDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}
	return nil
}
