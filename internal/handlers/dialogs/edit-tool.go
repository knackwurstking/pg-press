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
		if env.Verbose {
			h.Logger.Printf(env.ANSIVerbose+"Rendering edit tool dialog: %s%s"+env.ANSIReset, env.ANSIBlue, tool.String())
		}
	} else {
		t = templates.NewToolDialog()
		tName = "NewToolDialog"
		if env.Verbose {
			h.Logger.Printf(env.ANSIVerbose + "Rendering new tool dialog" + env.ANSIReset)
		}
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func (h *Handler) PostTool(c echo.Context) *echo.HTTPError {
	tool, merr := GetToolDialogForm(c)
	if merr != nil {
		return merr.Echo()
	}

	if env.Verbose {
		h.Logger.Printf(env.ANSIVerbose+"Creating new tool: %s%s"+env.ANSIReset, env.ANSIBlue, tool.String())
	}

	merr = h.DB.Tool.Tool.Create(tool)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// PutTool handles updating an existing tool
func (h *Handler) PutTool(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	tool, merr := GetToolDialogForm(c)
	if merr != nil {
		return merr.Echo()
	}
	tool.ID = toolID // Just to be sure

	if env.Verbose {
		h.Logger.Printf(env.ANSIVerbose+"Updating tool: %s"+env.ANSIReset, tool.String())
	}

	merr = h.DB.Tool.Tool.Update(tool)
	if merr != nil {
		return merr.Echo()
	}

	// TODO: Continue here...
	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for tool update", "error", merr)
	}

	// Set HX headers
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	utils.SetHXAfterSettle(c, map[string]any{
		"toolUpdated": map[string]string{
			"pageTitle": fmt.Sprintf("PG Presse | %s %s",
				tool.String(), tool.Position.GermanString()),
			"appBarTitle": fmt.Sprintf("%s %s", tool.String(),
				tool.Position.GermanString()),
		},
	})

	return nil
}

// TODO: This needs to be updated, ex.: the position needs to be replaced with the new slot logic
// TODO: Press selection needs to be kicked, the press page will handle that exclusively
func GetToolDialogForm(c echo.Context) (*shared.Tool, *errors.MasterError) {
	positionStr := c.FormValue("position")
	position := models.Position(positionStr)

	switch position {
	case models.PositionTop, models.PositionTopCassette, models.PositionBottom:
		// Valid position
	default:
		return nil, errors.NewMasterError(
			fmt.Errorf("invalid position: %s", positionStr),
			http.StatusBadRequest,
		)
	}

	data := &ToolDialogForm{Position: position}

	// Parse width
	if widthStr := c.FormValue("width"); widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return nil, errors.NewMasterError(err, http.StatusBadRequest)
		}
		data.Format.Width = width
	}

	// Parse height
	if heightStr := c.FormValue("height"); heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, errors.NewMasterError(err, http.StatusBadRequest)
		}
		data.Format.Height = height
	}

	// Parse type (with length limit)
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if len(data.Type) > 25 {
		return nil, errors.NewMasterError(
			fmt.Errorf("type must be 25 characters or less"),
			http.StatusBadRequest,
		)
	}

	// Parse code (required, with length limit)
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		return nil, errors.NewMasterError(
			fmt.Errorf("code is required"),
			http.StatusBadRequest,
		)
	}
	if len(data.Code) > 25 {
		return nil, errors.NewMasterError(
			fmt.Errorf("code must be 25 characters or less"),
			http.StatusBadRequest,
		)
	}

	return data, nil
}
