package shared

type PressUtilization struct {
	PressNumber       PressNumber `json:"press_number"`
	Type              PressType   `json:"type"`
	SlotUpper         *Tool       `json:"slot_upper"`
	SlotUpperCassette *Tool       `json:"slot_upper_cassette"`
	SlotLower         *Tool       `json:"slot_lower"`
}
