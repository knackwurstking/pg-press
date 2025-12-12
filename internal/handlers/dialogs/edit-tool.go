package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetEditTool(c echo.Context) error {
	var tool *models.Tool

	toolIDQuery, _ := utils.ParseQueryInt64(c, "id")
	if toolIDQuery > 0 {
		var merr *errors.MasterError
		tool, merr = h.registry.Tools.Get(models.ToolID(toolIDQuery))
		if merr != nil {
			return merr.Echo()
		}
	}

	var t templ.Component
	var tName string
	if tool != nil {
		t = templates.EditToolDialog(tool)
		tName = "EditToolDialog"
	} else {
		t = templates.NewToolDialog()
		tName = "NewToolDialog"
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func (h *Handler) PostEditTool(c echo.Context) error {
	slog.Info("Creating new tool")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	formData, merr := GetEditToolFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	tool := models.NewTool(formData.Position, formData.Format, formData.Code, formData.Type)
	tool.SetPress(formData.Press)

	_, merr = h.registry.Tools.Add(tool, user)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	title := "Neues Werkzeug erstellt"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for tool creation", "error", merr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditTool(c echo.Context) error {
	slog.Info("Updating existing tool")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	toolIDQuery, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(toolIDQuery)

	formData, merr := GetEditToolFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	tool.Press = formData.Press
	tool.Position = formData.Position
	tool.Format = formData.Format
	tool.Code = formData.Code
	tool.Type = formData.Type

	merr = h.registry.Tools.Update(tool, user)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	title := "Werkzeug aktualisiert"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

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

type EditToolFormData struct {
	Position models.Position
	Format   models.Format
	Type     string
	Code     string
	Press    *models.PressNumber
}

func GetEditToolFormData(c echo.Context) (*EditToolFormData, *errors.MasterError) {
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

	data := &EditToolFormData{Position: position}

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

	// Parse press number
	if pressStr := c.FormValue("press-selection"); pressStr != "" {
		press, err := strconv.Atoi(pressStr)
		if err != nil {
			return nil, errors.NewMasterError(
				fmt.Errorf("invalid press number: %v", err),
				http.StatusBadRequest,
			)
		}

		pressNumber := models.PressNumber(press)
		if !models.IsValidPressNumber(&pressNumber) {
			return nil, errors.NewMasterError(
				fmt.Errorf("invalid press number: must be 0, 2, 3, 4, or 5"),
				http.StatusBadRequest,
			)
		}
		data.Press = &pressNumber
	}

	return data, nil
}
