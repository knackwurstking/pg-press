package shared

type PressUtilization struct {
	PressNumber       PressNumber `json:"press_number"`
	CyclesOffset      int64       `json:"cycles_offset"`
	Type              MachineType `json:"type"`
	SlotUpper         *Tool       `json:"slot_upper"`
	SlotUpperCassette *Tool       `json:"slot_upper_cassette"`
	SlotLower         *Tool       `json:"slot_lower"`
}

func (pu *PressUtilization) Press() *Press {
	return &Press{
		ID:           pu.PressNumber,
		SlotUp:       pu.SlotUpper.ID,
		SlotDown:     pu.SlotLower.ID,
		CyclesOffset: pu.CyclesOffset,
		Type:         pu.Type,
	}
}
