package presscycle

import (
	"slices"
	"time"
)

type PressNumber int8

// IsValid checks if the press number is within the valid range (0-5)
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains([]PressNumber{0, 2, 3, 4, 5}, *n)
}

type Cycle struct {
	ID              int64       `json:"id"`
	PressNumber     PressNumber `json:"press_number"`
	SlotTop         int64       `json:"slot_top"`
	SlotTopCassette int64       `json:"slot_top_cassette"`
	SlotBottom      int64       `json:"slot_bottom"`
	Date            time.Time   `json:"date"`
	TotalCycles     int64       `json:"total_cycles"`
	PartialCycles   int64       `json:"partial_cycles"`
	PerformedBy     int64       `json:"performed_by"`
}

func NewCycle(press PressNumber, slotTop, slotTopCassette, slotBottom, totalCycles, user int64) *Cycle {
	return &Cycle{
		PressNumber:     press,
		SlotTop:         slotTop,
		SlotTopCassette: slotTopCassette,
		SlotBottom:      slotBottom,
		Date:            time.Now(),
		TotalCycles:     totalCycles,
		PerformedBy:     user,
	}
}

func NewPressCycleWithID(id int64, press PressNumber, slotTop, slotTopCassette, slotBottom, totalCycles, user int64, date time.Time) *Cycle {
	return &Cycle{
		ID:              id,
		PressNumber:     press,
		SlotTop:         slotTop,
		SlotTopCassette: slotTopCassette,
		SlotBottom:      slotBottom,
		Date:            date,
		TotalCycles:     totalCycles,
		PerformedBy:     user,
	}
}

func FilterSlots(slotTop, slotTopCassette, slotBottom int64, cycles ...*Cycle) []*Cycle {
	var filteredCycles []*Cycle

	for _, cycle := range cycles {
		if (slotTop > 0 && cycle.SlotTop == slotTop) ||
			(slotTopCassette > 0 && cycle.SlotTopCassette == slotTopCassette) ||
			(slotBottom > 0 && cycle.SlotBottom == slotBottom) {
			filteredCycles = append(filteredCycles, cycle)
		}
	}

	return filteredCycles
}
