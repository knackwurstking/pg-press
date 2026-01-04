package shared

import (
	"fmt"
)

const (
	PressNumber0 PressNumber = 0
	PressNumber2 PressNumber = 2
	PressNumber3 PressNumber = 3
	PressNumber4 PressNumber = 4
	PressNumber5 PressNumber = 5
)

var (
	AllPressNumbers []PressNumber = []PressNumber{
		PressNumber0,
		PressNumber2,
		PressNumber3,
		PressNumber4,
		PressNumber5,
	}
)

type PressNumber int8

func (p PressNumber) String() string {
	return fmt.Sprintf("%d", p)
}

func (p PressNumber) IsValid() bool {
	return p >= 0
}
