package dialogs

import (
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetToolDialog(c echo.Context) *echo.HTTPError {
	var tool *shared.Tool
	id, _ := shared.ParseQueryInt64(c, "id")
	if id > 0 {
		var merr *errors.MasterError
		tool, merr = db.GetTool(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
	}

	if tool != nil {
		log.Debug("Rendering edit tool dialog: %#v", tool.String())
		t := templates.EditToolDialog(tool)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolDialog")
		}
		return nil
	}

	log.Debug("Rendering new tool dialog...")
	t := templates.NewToolDialog()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewToolDialog")
	}

	return nil
}

func PostTool(c echo.Context) *echo.HTTPError {
	tool, verr := parseToolForm(c, nil)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	log.Debug("Creating new tool: %#v", tool.String())

	merr := db.AddTool(tool)
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
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	tool, verr := parseToolForm(c, tool)
	if verr != nil {
		return verr.MasterError().Echo()
	}

	log.Debug("Updating tool: %#v", tool.String())

	merr = db.UpdateTool(tool)
	if merr != nil {
		return merr.Echo()
	}

	// Set HX headers
	urlb.SetHXRedirect(c, urlb.UrlTool(tool.ID, 0, 0).Page)

	return nil
}

func parseToolForm(c echo.Context, tool *shared.Tool) (*shared.Tool, *errors.ValidationError) {
	if tool == nil {
		tool = &shared.Tool{}
	}
	tool.Type = strings.Trim(c.FormValue("type"), " ")
	tool.Code = strings.Trim(c.FormValue("code"), " ")

	// Need to convert the vPosition to an integer
	position, err := strconv.Atoi(c.FormValue("position"))
	if err != nil {
		return nil, errors.NewValidationError("invalid position: %d", position)
	}
	switch shared.Slot(position) {
	case shared.SlotUpper, shared.SlotLower:
		tool.Position = shared.Slot(position)
	default:
		return nil, errors.NewValidationError("invalid position: %d", position)
	}

	// Convert width and height to integers
	tool.Width, err = strconv.Atoi(c.FormValue("width"))
	if err != nil {
		return nil, errors.NewValidationError("invalid width: %s", c.FormValue("width"))
	}
	tool.Height, err = strconv.Atoi(c.FormValue("height"))
	if err != nil {
		return nil, errors.NewValidationError("invalid height: %s", c.FormValue("height"))
	}

	log.Debug("Tool dialog form values: tool=%v", tool)

	if verr := tool.Validate(); verr != nil {
		return tool, verr
	}

	return tool, nil
}
