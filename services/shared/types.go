package shared

import (
	"fmt"
)

type (
	EntityID   int64
	UnixMilly  int64
	TelegramID int64
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
