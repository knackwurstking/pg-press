package helper

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLGetPressNumberForTool string = `
		SELECT id, slot_up, slot_down 
		FROM presses 
		WHERE slot_up = :slot_up OR slot_down = :slot_down;
	`
)

func GetPressNumberForTool(db *common.DB, toolID shared.EntityID) (shared.PressNumber, shared.Slot) {
	var pressNumber shared.PressNumber = -1
	var slotUp, slotDown shared.EntityID

	_ = db.Press.Press.DB().QueryRow(SQLGetPressNumberForTool, toolID, toolID).Scan(&pressNumber, &slotUp, &slotDown)

	switch toolID {
	case slotUp:
		return pressNumber, shared.SlotUpper
	case slotDown:
		return pressNumber, shared.SlotLower
	default:
		return pressNumber, shared.SlotUnknown
	}
}

const (
	SQLListCyclesForPress string = `
		SELECT id, press_number, cycles, start, stop
		FROM press_cycles
		WHERE slot_up = :tool_id OR slot_down = :tool_id;
	`
)

// ListCyclesForTool returns all cycles associated with a specific tool by finding
// the press the tool is associated with and returning cycles for that press
func ListCyclesForTool(db *common.DB, toolID shared.EntityID) ([]*shared.Cycle, *errors.MasterError) {
	rows, err := db.Press.Cycle.DB().Query(SQLListCyclesForPress, toolID)
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
			&c.Cycles,
			&c.Start,
			&c.Stop,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		cycles = append(cycles, c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return cycles, nil
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
		if merr != nil {
			return nil, merr
		}
		utilizations[pn] = pu
	}

	return utilizations, nil
}
