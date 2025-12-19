package helper

import (
	"database/sql"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLGetToolIDForCassette string = `
		SELECT id FROM tools WHERE cassette = ?
	`

	SQLGetPressNumberForTool string = `
		SELECT id FROM presses WHERE slot_up = :tool_id OR slot_down = :tool_id
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "presses"
// -----------------------------------------------------------------------------

func GetPressNumberForTool(db *common.DB, toolID shared.EntityID) shared.PressNumber {
	var pressNumber shared.PressNumber = -1

	// Get the tool ID from the cassette ID, if the `toolID` is a cassette
	var id shared.EntityID
	db.Tool.Tools.DB().QueryRow(SQLGetToolIDForCassette, toolID).Scan(&id)
	if id > 0 {
		toolID = id
	}

	db.Press.Presses.DB().QueryRow(
		SQLGetPressNumberForTool, sql.Named("tool_id", toolID),
	).Scan(&pressNumber)

	return pressNumber
}

func GetPressUtilization(db *common.DB, pressNumber shared.PressNumber) (
	*shared.PressUtilization, *errors.MasterError,
) {
	pu := &shared.PressUtilization{PressNumber: pressNumber}

	press, merr := db.Press.Presses.GetByID(pressNumber)
	if merr != nil {
		return nil, merr
	}

	if press.SlotUp > 0 {
		// Get the top tool and cassette
		tool, merr := db.Tool.Tools.GetByID(press.SlotUp)
		if merr != nil {
			return nil, merr
		}
		pu.SlotUpper = tool

		if tool.Cassette > 0 {
			cassette, merr := db.Tool.Cassettes.GetByID(tool.Cassette)
			if merr != nil {
				return nil, merr
			}
			pu.SlotUpperCassette = cassette
		}
	} else {
		// Get the bottom tool
		pu.SlotUpper = nil
	}

	if press.SlotDown > 0 {
		tool, merr := db.Tool.Tools.GetByID(press.SlotDown)
		if merr != nil {
			return nil, merr
		}
		pu.SlotLower = tool
	} else {
		pu.SlotLower = nil
	}

	return pu, nil
}

func GetPressUtilizations(db *common.DB, pressNumbers []shared.PressNumber) (
	map[shared.PressNumber]*shared.PressUtilization, *errors.MasterError,
) {
	utilizations := make(map[shared.PressNumber]*shared.PressUtilization, len(pressNumbers))

	for _, pn := range pressNumbers {
		pu, merr := GetPressUtilization(db, pn)
		if merr != nil && merr.Code != http.StatusNotFound {
			return nil, merr
		}
		utilizations[pn] = pu
	}

	return utilizations, nil
}

// -----------------------------------------------------------------------------
// Table Helpers: "press_cycles"
// -----------------------------------------------------------------------------

func ListCyclesForTool(db *common.DB, toolID shared.EntityID) ([]*shared.Cycle, *errors.MasterError) {
	cycles, merr := db.Press.Cycles.List()
	if merr != nil {
		return nil, merr
	}

	// Filter cycles for the toolID
	i := 0
	for _, c := range cycles {
		if c.ToolID != toolID {
			continue
		}

		cycles[i] = c
		i++
	}

	return cycles[:i], nil
}

// TODO: Need to check for tool regenerations and only count cycles since the last regeneration
func GetTotalCyclesForTool(db *common.DB, toolID shared.EntityID) (int64, *errors.MasterError) {
	var totalCycles int64 = 0

	cycles, merr := db.Press.Cycles.List()
	if merr != nil {
		return totalCycles, merr
	}
	for _, c := range cycles {
		if c.ToolID != toolID {
			continue
		}

		totalCycles += c.PartialCycles
	}

	return totalCycles, nil
}
