package shared

import (
	"fmt"
)

const (
	EditorTypeTroubleReport EditorType = "troublereport"

	PressTypeSACMI = "SACMI"
	PressTypeSITI  = "SITI"

	SlotUnknown Slot = 0
	SlotUp      Slot = 1
	SlotDown    Slot = 2
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
	case SlotUp:
		return "UP"
	case SlotDown:
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
	case SlotUp:
		return "Oben"
	case SlotDown:
		return "Unten"
	default:
		return "?"
	}
}
