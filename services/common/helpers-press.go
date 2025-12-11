package common

import (
	"github.com/knackwurstking/pg-press/services/shared"
)

const (
	SQLGetPressNumberForTool string = `
		SELECT id, slot_up, slot_down FROM presses WHERE slot_up = :slot_up OR slot_down = :slot_down;
	`
)

func GetPressNumberForTool(db DB, toolID shared.EntityID) (shared.PressNumber, shared.Slot, error) {
	var pressNumber shared.PressNumber = -1
	var slotUp, slotDown shared.EntityID

	err := db.Press.Press.DB().QueryRow(SQLGetPressNumberForTool, toolID, toolID).Scan(&pressNumber, &slotUp, &slotDown)
	if err != nil {
		return pressNumber, shared.SlotUnknown, err
	}

	switch toolID {
	case slotUp:
		return pressNumber, shared.SlotUp, nil
	case slotDown:
		return pressNumber, shared.SlotDown, nil
	default:
		return pressNumber, shared.SlotUnknown, nil
	}
}
