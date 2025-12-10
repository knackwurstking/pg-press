package shared

type Cycle struct {
	PressNumber PressNumber `json:"press_number"` // PressNumber indicates which press machine performed the cycle
	Cycles      int64       `json:"cycles"`       // Cycles indicates the number of cycles performed in this record
	// TODO: Add date sepcific fields
}

// TODO: Add missing validate and clone methods to fit the Entity interface

var _ Entity[*Cycle] = (*Cycle)(nil)
