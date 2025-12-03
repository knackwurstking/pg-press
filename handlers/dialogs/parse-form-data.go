package dialogs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/labstack/echo/v4"
)

type DialogEditCycleFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *models.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
	ToolID       *models.ToolID // ToolID form field name "tool_id" (for tool change mode)
}

type DialogEditToolFormData struct {
	Position models.Position
	Format   models.Format
	Type     string
	Code     string
	Press    *models.PressNumber
}

type DialogEditToolRegenerationFormData struct {
	Reason string
}

func getEditCycleFormData(c echo.Context) (*DialogEditCycleFormData, error) {
	form := &DialogEditCycleFormData{}

	// Parse press number
	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, err
		}
		pn := models.PressNumber(press)
		form.PressNumber = &pn
	}

	// Parse date
	if dateString := c.FormValue("original_date"); dateString != "" {
		var err error
		form.Date, err = time.Parse(env.DateFormat, dateString)
		if err != nil {
			return nil, err
		}
	} else {
		form.Date = time.Now()
	}

	// Parse total cycles (required)
	totalCyclesString := c.FormValue("total_cycles")
	if totalCyclesString == "" {
		return nil, fmt.Errorf("form value total_cycles is required")
	}
	var err error
	form.TotalCycles, err = strconv.ParseInt(totalCyclesString, 10, 64)
	if err != nil {
		return nil, err
	}

	// Parse regenerating flag
	form.Regenerating = c.FormValue("regenerating") != ""

	// Parse tool_id if present (for tool change mode)
	if toolIDString := c.FormValue("tool_id"); toolIDString != "" {
		toolIDParsed, err := strconv.ParseInt(toolIDString, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid tool_id: %v", err)
		}
		toolID := models.ToolID(toolIDParsed)
		form.ToolID = &toolID
	}

	return form, nil
}

func getEditToolFormData(c echo.Context) (*DialogEditToolFormData, error) {
	positionStr := c.FormValue("position")
	position := models.Position(positionStr)

	switch position {
	case models.PositionTop, models.PositionTopCassette, models.PositionBottom:
		// Valid position
	default:
		return nil, errors.NewValidationError(fmt.Sprintf("invalid position: %s", positionStr))
	}

	data := &DialogEditToolFormData{Position: position}

	// Parse width
	if widthStr := c.FormValue("width"); widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid width: %v", err))
		}
		data.Format.Width = width
	}

	// Parse height
	if heightStr := c.FormValue("height"); heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid height: %v", err))
		}
		data.Format.Height = height
	}

	// Parse type (with length limit)
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if len(data.Type) > 25 {
		return nil, errors.NewValidationError("type must be 25 characters or less")
	}

	// Parse code (required, with length limit)
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		return nil, errors.NewValidationError("code is required")
	}
	if len(data.Code) > 25 {
		return nil, errors.NewValidationError("code must be 25 characters or less")
	}

	// Parse press number
	if pressStr := c.FormValue("press-selection"); pressStr != "" {
		press, err := strconv.Atoi(pressStr)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid press number: %v", err))
		}

		pressNumber := models.PressNumber(press)
		if !models.IsValidPressNumber(&pressNumber) {
			return nil, errors.NewValidationError("invalid press number: must be 0, 2, 3, 4, or 5")
		}
		data.Press = &pressNumber
	}

	return data, nil
}

func getMetalSheetFormData(c echo.Context) (*models.MetalSheet, error) {
	metalSheet := &models.MetalSheet{}

	// Parse required tile height field
	tileHeight, err := strconv.ParseFloat(c.FormValue("tile_height"), 64)
	if err != nil {
		return nil, err
	}
	metalSheet.TileHeight = tileHeight

	// Parse required value field
	value, err := strconv.ParseFloat(c.FormValue("value"), 64)
	if err != nil {
		return nil, err
	}
	metalSheet.Value = value

	// Parse optional marke height field
	if markeHeightStr := c.FormValue("marke_height"); markeHeightStr != "" {
		if markeHeight, err := strconv.Atoi(markeHeightStr); err == nil {
			metalSheet.MarkeHeight = markeHeight
		}
	}

	// Parse optional STF field
	if stfStr := c.FormValue("stf"); stfStr != "" {
		if stf, err := strconv.ParseFloat(stfStr, 64); err == nil {
			metalSheet.STF = stf
		}
	}

	// Parse optional STF Max field
	if stfMaxStr := c.FormValue("stf_max"); stfMaxStr != "" {
		if stfMax, err := strconv.ParseFloat(stfMaxStr, 64); err == nil {
			metalSheet.STFMax = stfMax
		}
	}

	// Parse identifier field with validation
	identifierStr := c.FormValue("identifier")
	if machineType, err := models.ParseMachineType(identifierStr); err == nil {
		metalSheet.Identifier = machineType
	} else {
		// Log the invalid value but don't fail - default to SACMI
		metalSheet.Identifier = models.MachineTypeSACMI // Default to SACMI
	}

	return metalSheet, nil
}

func getNoteFromFormData(c echo.Context) (note *models.Note, err error) {
	note = &models.Note{}

	// Parse level (required)
	levelStr := c.FormValue("level")
	if levelStr == "" {
		return nil, fmt.Errorf("level is required")
	}

	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		return nil, fmt.Errorf("invalid level format: %v", err)
	}

	// Validate level is within valid range (0=INFO, 1=ATTENTION, 2=BROKEN)
	if levelInt < 0 || levelInt > 2 {
		return nil, fmt.Errorf("invalid level value: %d (must be 0, 1, or 2)", levelInt)
	}

	note.Level = models.Level(levelInt)

	// Parse content (required)
	note.Content = strings.TrimSpace(c.FormValue("content"))
	if note.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Handle linked field - get first linked_tables value or empty string
	linkedTables := c.Request().Form["linked_tables"]
	if len(linkedTables) > 0 {
		note.Linked = linkedTables[0]
	}

	return note, nil
}

func getEditRegenerationFormData(c echo.Context) *DialogEditToolRegenerationFormData {
	return &DialogEditToolRegenerationFormData{
		Reason: c.FormValue("reason"),
	}
}
