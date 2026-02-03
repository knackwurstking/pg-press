package umbau

import (
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

	id, herr := utils.GetParamInt64(c, "press")
	if herr != nil {
		return herr.Echo()
	}
	press, herr := db.GetPress(shared.EntityID(id))
	if herr != nil {
		return herr.Echo()
	}

	u, merr := db.GetPressUtilization(press.ID)
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := db.ListTools()
	if merr != nil {
		return merr.Echo()
	}
	upperTools := make([]*shared.Tool, 0)
	lowerTools := make([]*shared.Tool, 0)
	for _, t := range tools {
		switch t.Position {
		case shared.SlotUpper:
			upperTools = append(upperTools, t)
		case shared.SlotLower:
			lowerTools = append(lowerTools, t)
		}
	}

	t := templates.Page(&templates.PageProps{
		Press:            press,
		User:             user,
		CurrentUpperTool: u.SlotUpper,
		CurrentLowerTool: u.SlotLower,
		UpperTools:       upperTools,
		LowerTools:       lowerTools,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Umbau Page")
	}

	return nil
}

func PostUmbauPage(c echo.Context) *echo.HTTPError {
	id, herr := utils.GetParamInt64(c, "press")
	if herr != nil {
		return herr.Echo()
	}
	pressID := shared.EntityID(id)

	data, eerr := getFormData(c)
	if eerr != nil {
		return eerr
	}

	// Press Utilization
	u, merr := db.GetPressUtilization(pressID)
	if merr != nil {
		return merr.WrapEcho("get press utilization")
	}

	// Old Tools for setting cycles
	tools := make([]*shared.Tool, 0)
	if u.SlotUpper != nil {
		tools = append(tools, u.SlotLower)
	}
	if u.SlotUpperCassette != nil {
		tools = append(tools, u.SlotUpperCassette)
	}
	if u.SlotLower != nil {
		tools = append(tools, u.SlotLower)
	}
	// Set cycles for old tools
	for _, t := range tools {
		merr = db.AddCycle(
			shared.NewCycle(
				t.ID,
				pressID,
				data.totalCycles,
				shared.NewUnixMilli(time.Now()),
			),
		)
		if merr != nil {
			return merr.WrapEcho("add cycle")
		}
	}

	// Update press with new tools
	press := u.Press()
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
