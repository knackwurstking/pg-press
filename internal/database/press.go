package database

import "time"

type PressNumber int8

type Press struct {
	Number        PressNumber           `json:"number"` // Number of the press, 0-5 (Unique)
	From          time.Time             `json:"from"`
	To            *time.Time            `json:"to"`
	TotalCycles   int64                 `json:"total_cycles"`
	PartialCycles int64                 `json:"partial_cycles"`
	Mods          []*Modified[PressMod] `json:"mods"`
}

func NewPress(number PressNumber, from time.Time, to *time.Time, total, partial int64, m ...*Modified[PressMod]) *Press {
	return &Press{
		Number:        number,
		From:          from,
		To:            to,
		TotalCycles:   total,
		PartialCycles: partial,
	}
}

type PressMod struct {
	From          time.Time  `json:"from"`
	To            *time.Time `json:"to"`
	TotalCycles   int64      `json:"total_cycles"`
	PartialCycles int64      `json:"partial_cycles"`
}
