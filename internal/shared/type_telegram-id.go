package shared

import "fmt"

type TelegramID int64

func (id TelegramID) String() string {
	return fmt.Sprintf("%d", id)
}
