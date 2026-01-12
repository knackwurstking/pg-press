package umbau

import (
	"fmt"
	"net/http"
	"strconv"

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

	var pressNumber shared.PressNumber
	if pn, merr := utils.GetParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		if pressNumber = shared.PressNumber(pn); !pressNumber.IsValid() {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Sprintf("invalid press number: %d", pressNumber),
			)
		}
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

// TODO: ...
func PostUmbauPage(c echo.Context) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	var pressNumber shared.PressNumber
	if pn, merr := utils.GetParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		pressNumber = shared.PressNumber(pn)
	}
	if !pressNumber.IsValid() {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"missing total cycles",
		)
	}
	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid total cycles",
		)
	}

	id, err := strconv.ParseInt(c.FormValue("top"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"missing top tool",
		)
	}
	topTool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	id, err = strconv.ParseInt(c.FormValue("bottom"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"missing bottom tool",
		)
	}
	bottomTool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	if topTool.Format.String() != bottomTool.Format.String() {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"tools are not compatible",
		)
	}

	currentTools, merr := db.GetToolsByPress(shared.PressNumber(pressNumber))
	if merr != nil {
		return merr.Echo()
	}

	for _, tool := range currentTools {
		cycle := shared.NewCycle(shared.PressNumber(pressNumber), tool.ID, tool.Position, totalCycles, user.TelegramID)

		_, merr = db.AddPressCycle(
			cycle.PressNumber, cycle.ToolID, cycle.ToolPosition, cycle.TotalCycles, cycle.PerformedBy,
		)
		if merr != nil {
			return merr.Echo()
		}
	}

	for _, tool := range currentTools {
		merr := db.UpdateToolPress(tool.ID, nil)
		if merr != nil {
			return merr.Echo()
		}
	}

	newTools := []*shared.Tool{topTool, bottomTool}
	for _, tool := range newTools {
		merr := db.UpdateToolPress(tool.ID, &shared.PressNumber(pressNumber))
		if merr != nil {
			return merr.Echo()
		}
	}

	title := fmt.Sprintf("Werkzeugwechsel Presse %d", pressNumber)
	content := fmt.Sprintf(
		"Umbau abgeschlossen für Presse %d.\nEingebautes Oberteil: %s\nEingebautes Unterteil: %s\nGesamtzyklen: %d",
		pressNumber, topTool.String(), bottomTool.String(), totalCycles,
	)

	merr = db.AddFeedEntry(title, content, user.TelegramID)
	if merr != nil {
		return merr.Echo()
	}

	return nil
}

func findToolByID(tools []*shared.Tool, toolID shared.ToolID) (*shared.Tool, *errors.MasterError) {
	for _, tool := range tools {
		if tool.ID == toolID {
			return tool, nil
		}
	}
	return nil, errors.NewMasterError(
		fmt.Errorf("tool not found: %d", toolID),
		http.StatusBadRequest,
	)
}
