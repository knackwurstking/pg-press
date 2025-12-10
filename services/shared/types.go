package shared

import (
	"fmt"
)

const (
	EditorTypeTroubleReport EditorType = "troublereport"

	PressTypeSACMI = "SACMI"
	PressTypeSITI  = "SITI"
)

type (
	EntityID    int64
	TelegramID  int64
	EditorType  string
	PressNumber int8
	PressType   string
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

func (p PressNumber) IsValid() bool {
	switch p {
	case 0, 2, 3, 4, 5:
		return true
	default:
		return false
	}
}
