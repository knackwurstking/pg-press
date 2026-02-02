package shared

type PressUtilization struct {
	PressNumber       PressNumber `json:"press_number"`
	Type              MachineType `json:"type"`
	Code              string      `json:"code"`
	CyclesOffset      int64       `json:"cycles_offset"`
	SlotUpper         *Tool       `json:"slot_upper"`
	SlotUpperCassette *Tool       `json:"slot_upper_cassette"`
	SlotLower         *Tool       `json:"slot_lower"`
}

func (pu *PressUtilization) Press() *Press {
	return &Press{
		ID:           pu.PressNumber,
		Type:         pu.Type,
		Code:         pu.Code,
		SlotUp:       pu.SlotUpper.ID,
		SlotDown:     pu.SlotLower.ID,
		CyclesOffset: pu.CyclesOffset,
	}
}
