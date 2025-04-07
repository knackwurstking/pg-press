package types

import "fmt"

type ModifiedByData interface {
	MetalSheet | MetalSheetPress | MetalSheetData
}

type ModifiedBy[T ModifiedByData] struct {
	Time   int    `json:"time"`
	User   string `json:"user"`
	UserID int    `json:"user_id"`
	Data   T      `json:"data"`
}

func (mb *ModifiedBy[T]) Key() string {
	return fmt.Sprintf("%d:%s:%d", mb.Time, mb.User, mb.UserID)
}

func (mb *ModifiedBy[T]) GetData() *T {
	return &mb.Data
}

type Modifiers[T ModifiedByData] []ModifiedBy[T]

func (m *Modifiers[T]) Add(mb *ModifiedBy[T]) {
	*m = append(*m, *mb)
}
