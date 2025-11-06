package dialogs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/labstack/echo/v4"
)

type DialogCycleEditFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *models.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
	ToolID       *models.ToolID // ToolID form field name "tool_id" (for tool change mode)
}

func getCycleFormData(c echo.Context) (*DialogCycleEditFormData, error) {
	form := &DialogCycleEditFormData{}

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
