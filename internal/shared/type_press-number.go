package shared

import (
	"fmt"
)

type PressNumber int8

func (p PressNumber) String() string {
	return fmt.Sprintf("%d", p)
}

func (p PressNumber) IsValid() bool {
	return p >= 0
}
