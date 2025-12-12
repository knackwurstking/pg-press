package helper

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLGetPressNumberForTool string = `
		SELECT id, slot_up, slot_down FROM presses WHERE slot_up = :slot_up OR slot_down = :slot_down;
	`
)

func GetPressNumberForTool(db *common.DB, toolID shared.EntityID) (shared.PressNumber, shared.Slot, *errors.MasterError) {
	var pressNumber shared.PressNumber = -1
	var slotUp, slotDown shared.EntityID

	err := db.Press.Press.DB().QueryRow(SQLGetPressNumberForTool, toolID, toolID).Scan(&pressNumber, &slotUp, &slotDown)
	if err != nil {
		return pressNumber, shared.SlotUnknown, errors.NewMasterError(err, 0)
	}

	switch toolID {
	case slotUp:
		return pressNumber, shared.SlotPressUp, nil
	case slotDown:
		return pressNumber, shared.SlotPressDown, nil
	default:
		return pressNumber, shared.SlotUnknown, nil
	}
}
