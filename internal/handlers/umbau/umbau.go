package umbau

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/umbau/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetUmbauPage(c echo.Context) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	pressNumber, eerr := getParamPressNumber(c)
	if eerr != nil {
		return eerr
	}

	tools, merr := db.ListTools()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(&templates.PageProps{
		PressNumber: pressNumber,
		User:        user,
		Tools:       tools,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Umbau Page")
	}

	return nil
}

func PostUmbauPage(c echo.Context) *echo.HTTPError {
	pressNumber, eerr := getParamPressNumber(c)
	if eerr != nil {
		return eerr
	}

	data, eerr := getFormData(c)
	if eerr != nil {
		return eerr
	}

	// Press Utilization
	pu, merr := db.GetPressUtilizations(pressNumber)
	if merr != nil {
		return merr.WrapEcho("get press utilization")
	}
	pressUtilization, ok := pu[pressNumber]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("no data for press %d", pressNumber))
	}

	// Old Tools for setting cycles
	tools := make([]*shared.Tool, 0)
	if pressUtilization.SlotUpper != nil {
		tools = append(tools, pressUtilization.SlotLower)
	}
	if pressUtilization.SlotUpperCassette != nil {
		tools = append(tools, pressUtilization.SlotUpperCassette)
	}
	if pressUtilization.SlotLower != nil {
		tools = append(tools, pressUtilization.SlotLower)
	}
	// Set cycles for old tools
	for _, t := range tools {
		merr = db.AddCycle(&shared.Cycle{
			ToolID:      t.ID,
			PressNumber: pressNumber,
			PressCycles: data.totalCycles,
			Stop:        shared.NewUnixMilli(time.Now()),
		})
		if merr != nil {
			return merr.WrapEcho("add cycle")
		}
	}

	// Update press with new tools
	press := pressUtilization.Press()
	tools = []*shared.Tool{data.upperTool, data.lowerTool}
	for _, t := range tools {
		switch t.Position {
		case shared.SlotUpper:
			press.SlotUp = t.ID
		case shared.SlotLower:
			press.SlotDown = t.ID
		}
	}
	merr = db.UpdatePress(press)
	if merr != nil {
		return merr.WrapEcho("update press")
	}

	return nil
}

func getParamPressNumber(c echo.Context) (shared.PressNumber, *echo.HTTPError) {
	// PressNumber from param
	var pressNumber shared.PressNumber
	pn, merr := utils.GetParamInt8(c, "press")
	if merr != nil {
		return pressNumber, merr.Echo()
	}
	pressNumber = shared.PressNumber(pn)
	if !pressNumber.IsValid() {
		return pressNumber, echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	return pressNumber, nil
}

type formData struct {
	totalCycles int64
	upperTool   *shared.Tool
	lowerTool   *shared.Tool
}

func getFormData(c echo.Context) (*formData, *echo.HTTPError) {
	data := &formData{}

	// TotalCycles from form
	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return data, echo.NewHTTPError(http.StatusBadRequest, "missing total cycles")
	}
	var err error
	data.totalCycles, err = strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return data, echo.NewHTTPError(http.StatusBadRequest, "invalid total cycles")
	}

	// Top Tool from form
	id, err := strconv.ParseInt(c.FormValue("top"), 10, 64)
	if err != nil {
		return data, echo.NewHTTPError(http.StatusBadRequest, "missing top tool")
	}
	var merr *errors.HTTPError
	data.upperTool, merr = db.GetTool(shared.EntityID(id))
	if merr != nil {
		return data, merr.WrapEcho("get (upper) tool with ID %d", id)
	}

	// Bottom tool from form
	id, err = strconv.ParseInt(c.FormValue("bottom"), 10, 64)
	if err != nil {
		return data, echo.NewHTTPError(http.StatusBadRequest, "missing bottom tool")
	}
	data.lowerTool, merr = db.GetTool(shared.EntityID(id))
	if merr != nil {
		return data, merr.WrapEcho("get (bottom) tool with ID %d", id)
	}

	if data.upperTool.Width == data.lowerTool.Width && data.upperTool.Height == data.lowerTool.Height {
		return data, echo.NewHTTPError(http.StatusBadRequest, "incompatible tools format")
	}

	return data, nil
}
