package shared

type PressUtilization struct {
	PressNumber       PressNumber `json:"press_number"`
	SlotUpper         *Tool       `json:"slot_upper"`
	SlotUpperCassette *Cassette   `json:"slot_upper_cassette"`
	SlotLower         *Tool       `json:"slot_lower"`
}
