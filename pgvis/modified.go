// ai: Organize
package pgvis

import (
	"fmt"
	"time"
)

// Modified represents a modification record that tracks changes made to any type T
// It stores information about who made the change, when it was made, and the original value
type Modified[T any] struct {
	User     *User `json:"user"`     // The user who made the modification
	Time     int64 `json:"time"`     // UNIX millisecond timestamp of the modification
	Original T     `json:"original"` // The original value before modification
}

// NewModified creates a new modification record with the current timestamp
// It records the user who made the change and the original value being modified
func NewModified[T any](user *User, original T) *Modified[T] {
	if user == nil {
		// Create a default user for system modifications
		user = &User{
			UserName: "system",
		}
	}

	return &Modified[T]{
		User:     user,
		Time:     time.Now().UnixMilli(),
		Original: original,
	}
}

// NewModifiedWithTime creates a new modification record with a specific timestamp
// Useful for importing historical data or testing
func NewModifiedWithTime[T any](user *User, original T, timestamp int64) *Modified[T] {
	if user == nil {
		user = &User{
			UserName: "system",
		}
	}

	return &Modified[T]{
		User:     user,
		Time:     timestamp,
		Original: original,
	}
}

// GetTime returns the modification time as a Go time.Time
func (m *Modified[T]) GetTime() time.Time {
	return time.UnixMilli(m.Time)
}

// GetUserName returns the username of the user who made the modification
// Returns "unknown" if the user is nil
func (m *Modified[T]) GetUserName() string {
	if m.User == nil {
		return "unknown"
	}
	return m.User.UserName
}

// IsModifiedBy checks if the modification was made by a specific user
func (m *Modified[T]) IsModifiedBy(userName string) bool {
	if m.User == nil {
		return userName == "unknown" || userName == ""
	}
	return m.User.UserName == userName
}

// String returns a human-readable representation of the modification
func (m *Modified[T]) String() string {
	userName := m.GetUserName()
	timeStr := m.GetTime().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("Modified by %s at %s", userName, timeStr)
}

// Age returns the duration since the modification was made
func (m *Modified[T]) Age() time.Duration {
	return time.Since(m.GetTime())
}

// IsOlderThan checks if the modification is older than the specified duration
func (m *Modified[T]) IsOlderThan(duration time.Duration) bool {
	return m.Age() > duration
}

// IsNewerThan checks if the modification is newer than the specified duration
func (m *Modified[T]) IsNewerThan(duration time.Duration) bool {
	return m.Age() < duration
}

// Validate checks if the modification record is valid
func (m *Modified[T]) Validate() error {
	if m.Time <= 0 {
		return fmt.Errorf("modification time must be positive")
	}

	if m.Time > time.Now().UnixMilli() {
		return fmt.Errorf("modification time cannot be in the future")
	}

	return nil
}
