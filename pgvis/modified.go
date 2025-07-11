package pgvis

import (
	"encoding/json"
	"fmt"
)

type Modified[T any] struct {
	User     *User `json:"user"`
	Time     int64 `json:"time"`
	Original T     `json:"original"`
}

func NewModified[T any](user *User) *Modified[T] {
	return &Modified[T]{
		User: user,
	}
}

func (m *Modified[T]) JSON() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Errorf("marshal modified failed: %s", err.Error()))
	}

	return data
}
