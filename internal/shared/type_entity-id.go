package shared

import "fmt"

type EntityID int64

func (id EntityID) String() string {
	return fmt.Sprintf("%d", id)
}
