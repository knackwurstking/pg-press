package shared

import (
	"fmt"
)

const (
	EditorTypeTroubleReport EditorType = "troublereport"

	PressTypeSACMI = "SACMI"
	PressTypeSITI  = "SITI"

	SlotUnknown           Slot = 0
	SlotPressUp           Slot = 1
	SlotPressDown         Slot = 2
	SlotUpperToolCassette Slot = 10
	//SlotLowerToolCassette Slot = 20
)

type (
	EntityID    int64
	TelegramID  int64
	EditorType  string
	PressNumber int8
	PressType   string
	Slot        int
)

func (id EntityID) String() string {
	return fmt.Sprintf("%d", id)
}

func (id TelegramID) String() string {
	return fmt.Sprintf("%d", id)
}

func (p PressNumber) String() string {
	return fmt.Sprintf("%d", p)
}

func (p Slot) String() string {
	switch p {
	case SlotPressUp:
		return "UP"
	case SlotPressDown:
		return "DOWN"
	default:
		return "UNKNOWN"
	}
}

func (p PressNumber) IsValid() bool {
	switch p {
	case 0, 2, 3, 4, 5:
		return true
	default:
		return false
	}
}

func (p Slot) German() string {
	switch p {
	case SlotPressUp:
		return "Oben"
	case SlotPressDown:
		return "Unten"
	default:
		return "?"
	}
}
