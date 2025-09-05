package models

import "time"

const (
	MinPressNumber = 0
	MaxPressNumber = 5
)

type PressNumber int8

// IsValid checks if the press number is within the valid range (0-5)
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return *n >= MinPressNumber && *n <= MaxPressNumber
}

type PressCycle struct {
	ID              int64       `json:"id"`
	PressNumber     PressNumber `json:"press_number"` // PressNumber is optional
	SlotTop         int64       `json:"slot_top"`
	SlotTopCassette int64       `json:"slot_top_cassette"`
	SlotBottom      int64       `json:"slot_bottom"`
	Date            time.Time   `json:"date"`
	TotalCycles     int64       `json:"total_cycles"`
	PerformedBy     int64       `json:"performed_by"`
}

// TODO: Need to create a new ID type sometime
func NewPressCycle(slotTop, slotTopCassette, slotBottom int64, press PressNumber, totalCycles, user int64) *PressCycle {
	return &PressCycle{
		SlotTop:         slotTop,
		SlotTopCassette: slotTopCassette,
		SlotBottom:      slotBottom,
		PressNumber:     press,
		Date:            time.Now(),
		TotalCycles:     totalCycles,
		PerformedBy:     user,
	}
}

func NewPressCycleWithID(id, slotTop, slotTopCassette, slotBottom int64, press PressNumber, totalCycles, user int64, date time.Time) *PressCycle {
	return &PressCycle{
		SlotTop:         slotTop,
		SlotTopCassette: slotTopCassette,
		SlotBottom:      slotBottom,
		ID:              id,
		PressNumber:     press,
		Date:            date,
		TotalCycles:     totalCycles,
		PerformedBy:     user,
	}
}
