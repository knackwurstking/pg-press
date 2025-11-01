// TODO: Remove useless stuff
package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type ModificationID int64

// ModificationType represents the type of entity being modified
type ModificationType string

const (
	ModificationTypeTroubleReport ModificationType = "trouble_reports"
	ModificationTypeMetalSheet    ModificationType = "metal_sheets"
	ModificationTypeTool          ModificationType = "tools"
	ModificationTypePressCycle    ModificationType = "press_cycles"
	ModificationTypeUser          ModificationType = "users"
	ModificationTypeNote          ModificationType = "notes"
	ModificationTypeAttachment    ModificationType = "attachments"
)

// Modification represents a modification in the database
type Modification[T any] struct {
	ID        ModificationID `json:"id"`
	UserID    TelegramID     `json:"user_id"`
	Data      []byte         `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

func NewModification[T any](data T, telegramID TelegramID) *Modification[T] {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	return &Modification[T]{
		UserID:    telegramID,
		Data:      jsonData,
		CreatedAt: time.Now(),
	}
}

// UnmarshalData unmarshals the data into the provided value
func (m *Modification[T]) UnmarshalData(v *T) error {
	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}
	return json.Unmarshal(m.Data, v)
}

// MarshalData marshals the data into JSON format and stores it in the Data field
func (m *Modification[T]) MarshalData(v T) ([]byte, error) {
	var err error
	m.Data, err = json.Marshal(v)
	return m.Data, err
}

// GetData unmarshals and returns the data as the specified type
func (m *Modification[T]) GetData() (*T, error) {
	var data T
	if err := m.UnmarshalData(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

// IsEmpty returns true if the modification has no data
func (m *Modification[T]) IsEmpty() bool {
	return len(m.Data) == 0
}

// Age returns the duration since the modification was created
func (m *Modification[T]) Age() time.Duration {
	return time.Since(m.CreatedAt)
}

// IsOlderThan returns true if the modification is older than the specified duration
func (m *Modification[T]) IsOlderThan(duration time.Duration) bool {
	return m.Age() > duration
}

// IsNewerThan returns true if the modification is newer than the specified duration
func (m *Modification[T]) IsNewerThan(duration time.Duration) bool {
	return m.Age() < duration
}

// String returns a string representation of the modification
func (m *Modification[T]) String() string {
	return fmt.Sprintf("Modification{ID: %d, UserID: %d, CreatedAt: %s}",
		m.ID, m.UserID, m.CreatedAt.Format("2006-01-02 15:04:05"))
}

// Validate performs basic validation on the modification
func (m *Modification[T]) Validate() error {
	if m.ID < 0 {
		return fmt.Errorf("modification ID cannot be negative")
	}
	if m.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if m.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}
	if m.CreatedAt.After(time.Now()) {
		return fmt.Errorf("created_at cannot be in the future")
	}
	return nil
}

// Clone creates a deep copy of the modification
func (m *Modification[T]) Clone() *Modification[T] {
	dataCopy := make([]byte, len(m.Data))
	copy(dataCopy, m.Data)

	return &Modification[T]{
		ID:        m.ID,
		UserID:    m.UserID,
		Data:      dataCopy,
		CreatedAt: m.CreatedAt,
	}
}
