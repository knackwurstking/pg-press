package shared

import (
	"fmt"
)

const (
	EditorTypeTroubleReport EditorType = "troublereport"
)

type (
	EntityID    int64
	UnixMilly   int64
	TelegramID  int64
	PressNumber int8
	EditorType  string
)

func (id EntityID) String() string {
	return fmt.Sprintf("%d", id)
}

func (u UnixMilly) String() string {
	return fmt.Sprintf("%d", u)
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
