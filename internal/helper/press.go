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
