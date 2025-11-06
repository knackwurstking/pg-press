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

func getEditCycleFormData(c echo.Context) (*DialogEditCycleFormData, error) {
	form := &DialogEditCycleFormData{}

	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, err
		}

		pn := models.PressNumber(press)
		form.PressNumber = &pn
	}

	if dateString := c.FormValue("original_date"); dateString != "" {
		var err error
		form.Date, err = time.Parse(env.DateFormat, dateString)
		if err != nil {
			return nil, err
		}
	} else {
		form.Date = time.Now()
	}

	if totalCyclesString := c.FormValue("total_cycles"); totalCyclesString == "" {
		return nil, fmt.Errorf("form value total_cycles is required")
	} else {
		var err error
		form.TotalCycles, err = strconv.ParseInt(totalCyclesString, 10, 64)
		if err != nil {
			return nil, err
		}
	}

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

	// Parse type
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if len(data.Type) > 25 {
		return nil, errors.NewValidationError("type must be 25 characters or less")
	}

	// Parse code
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		return nil, errors.NewValidationError("code is required")
	}
	if len(data.Code) > 25 {
		return nil, errors.NewValidationError("code must be 25 characters or less")
	}

	// Parse press
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
