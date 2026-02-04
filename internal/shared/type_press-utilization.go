package shared

type PressUtilization struct {
	PressID           EntityID    `json:"press_id"`
	PressNumber       PressNumber `json:"press_number"`
	PressType         MachineType `json:"type"`
	PressCode         string      `json:"code"`
	CyclesOffset      int64       `json:"cycles_offset"`
	SlotUpper         *Tool       `json:"slot_upper"`
	SlotUpperCassette *Tool       `json:"slot_upper_cassette"`
	SlotLower         *Tool       `json:"slot_lower"`
}

func (pu *PressUtilization) Press() *Press {
	return &Press{
		ID:           pu.PressID,
		Number:       pu.PressNumber,
		Type:         pu.PressType,
		Code:         pu.PressCode,
		SlotUp:       pu.SlotUpper.ID,
		SlotDown:     pu.SlotLower.ID,
		CyclesOffset: pu.CyclesOffset,
	}
}
