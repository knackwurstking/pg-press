package pgvis

import "time"

type Modified[T any] struct {
	User     *User `json:"user"`
	Time     int64 `json:"time"`
	Original T     `json:"original"`
}

func NewModified[T any](user *User, original T) *Modified[T] {
	return &Modified[T]{
		User: user,
		Time: time.Now().UnixMilli(),
		Original: original,
	}
}
