package types

import "fmt"

type ModifiedBy struct {
	Time   int    `json:"time"`
	User   string `json:"user"`
	UserID int    `json:"user_id"`
}

func (mb *ModifiedBy) Key() string {
	return fmt.Sprintf("%d:%s:%d", mb.Time, mb.User, mb.UserID)
}

func (mb *ModifiedBy) Is(k *ModifiedBy) bool {
	return k.Key() == mb.Key()
}

func (mb *ModifiedBy) IsKey(k string) bool {
	return k == mb.Key()
}
