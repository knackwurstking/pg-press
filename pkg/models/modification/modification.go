package modification

import (
	"encoding/json"
	"time"
)

// Modification represents a modification in the database
type Modification[T any] struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

// UnmarshalData unmarshals the data into the provided value
func (m *Modification[T]) UnmarshalData(v T) error {
	return json.Unmarshal(m.Data, v)
}

// MarshalData marshals the data into JSON format and stores it in the Data field
func (m *Modification[T]) MarshalData(v T) ([]byte, error) {
	var err error
	m.Data, err = json.Marshal(v)
	return m.Data, err
}
