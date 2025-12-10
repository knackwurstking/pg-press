package shared

type Cycle struct {
	PressNumber PressNumber `json:"press_number"`
	Cycles      int64       `json:"cycles"`
}

// TODO: Add missing validate and clone methods to fit the Entity interface

var _ Entity[*Cycle] = (*Cycle)(nil)
