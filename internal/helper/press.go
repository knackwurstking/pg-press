package helper

import (
	"database/sql"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/press"
	"github.com/knackwurstking/pg-press/internal/shared"
)

func GetPressNumberForTool(db *common.DB, toolID shared.EntityID) shared.PressNumber {
	var pressNumber shared.PressNumber = -1

	var id shared.EntityID
	db.Tool.Tool.DB().QueryRow(
		`SELECT id FROM tools WHERE cassette = ?`, toolID,
	).Scan(&id)
	if id > 0 {
		db.Press.Press.DB().QueryRow(
			`SELECT id FROM presses WHERE slot_up = :tool_id OR slot_down = :tool_id`,
			sql.Named("tool_id", id),
		).Scan(&pressNumber)
	}

	return pressNumber
}

// ListCyclesForTool returns all cycles associated with a specific tool by finding
// the press the tool is associated with and returning cycles for that press
func ListCyclesForTool(db *common.DB, toolID shared.EntityID) ([]*shared.Cycle, *errors.MasterError) {
	rows, err := db.Press.Cycle.DB().Query(
		`
			SELECT id, tool_id, position, press_number, cycles, start, stop
			FROM press_cycles
			WHERE slot_up = :tool_id OR slot_down = :tool_id
			ORDER BY press_number ASC, stop DESC;
		`,
		sql.Named("tool_id", toolID),
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	cycles := []*shared.Cycle{}
	for rows.Next() {
		c := &shared.Cycle{}
		err := rows.Scan(
			&c.ID,
			&c.PressNumber,
			&c.PressCycles,
			&c.Start,
			&c.Stop,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		c.PartialCycles = press.CalculatePartialCycles(db.Press.Cycle.DB(), c)
		cycles = append(cycles, c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return cycles, nil
}

func GetTotalCyclesForTool(db *common.DB, toolID shared.EntityID) (int64, *errors.MasterError) {
	var totalCycles int64 = 0

	rows, err := db.Press.Cycle.DB().Query(
		`
			SELECT cycles
			FROM press_cycles
			WHERE slot_up = :tool_id OR slot_down = :tool_id;
		`,
		sql.Named("tool_id", toolID),
	)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	for rows.Next() {
		var cycles int64
		err := rows.Scan(&cycles)
		if err != nil {
			return 0, errors.NewMasterError(err, 0)
		}
		totalCycles += cycles
	}

	if err = rows.Err(); err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return totalCycles, nil
}

func GetPressUtilization(db *common.DB, pressNumber shared.PressNumber) (
	*shared.PressUtilization, *errors.MasterError,
) {
	pu := &shared.PressUtilization{PressNumber: pressNumber}

	press, merr := db.Press.Press.GetByID(pressNumber)
	if merr != nil {
		return nil, merr
	}

	if press.SlotUp > 0 {
		// Get the top tool and cassette
		tool, merr := db.Tool.Tool.GetByID(press.SlotUp)
		if merr != nil {
			return nil, merr
		}
		pu.SlotUpper = tool

		if tool.Cassette > 0 {
			cassette, merr := db.Tool.Cassette.GetByID(tool.Cassette)
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
		tool, merr := db.Tool.Tool.GetByID(press.SlotDown)
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
