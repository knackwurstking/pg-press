package pgvis

import (
	"encoding/json"
	"time"
)

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

func (m *Modified[T]) ToJSON() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return data
}

func (m *Modified[T]) FromJSON(b []byte) []byte {
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return data
}
