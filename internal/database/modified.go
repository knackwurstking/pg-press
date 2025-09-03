package database

import (
	"fmt"
	"slices"
	"time"

	"github.com/knackwurstking/pgpress/internal/dberror"
)

type Mods[T any] []*Modified[T]

func (m *Mods[T]) Reversed() []*Modified[T] {
	reversed := make([]*Modified[T], len(*m))
	copy(reversed, *m)
	slices.Reverse(reversed)
	return reversed
}

func (m *Mods[T]) Add(user *User, data T) {
	*m = append(*m, NewModified(user, data))
}

func (m *Mods[T]) First() *Modified[T] {
	return (*m)[0]
}

func (m *Mods[T]) Current() *Modified[T] {
	if len(*m) == 0 {
		return nil
	}
	return (*m)[len(*m)-1]
}

func (m *Mods[T]) Get(time int64) (*Modified[T], error) {
	for _, mod := range *m {
		if mod.Time == time {
			return mod, nil
		}
	}
	return nil, dberror.ErrNotFound
}

//func (m *Mods[T]) Rollback(time int64) error {
//	for i, mod := range *m {
//		if mod.Time == time {
//			// Move the matching modification to the current position
//			before := (*m)[0:i]
//			restMods := append(before, (*m)[i+1:]...)
//			*m = append(restMods, mod)
//			return nil
//		}
//	}
//	return ErrNotFound
//}

// Modified represents a modification record that tracks changes made to any type T
type Modified[T any] struct {
	User *User `json:"user"`
	Time int64 `json:"time"` // Time of modification in milliseconds since Unix epoch,
	// should be unique
	Data T `json:"data"`
}

// NewModified creates a new modification record with the current timestamp
func NewModified[T any](user *User, data T) *Modified[T] {
	if user == nil {
		user = &User{UserName: "system"}
	}

	return &Modified[T]{
		User: user,
		Time: time.Now().UnixMilli(),
		Data: data,
	}
}

// GetTime returns the modification time as a Go time.Time
func (m *Modified[T]) GetTime() time.Time {
	return time.UnixMilli(m.Time)
}

func (m *Modified[T]) GetTimeString() string {
	return m.GetTime().Format("2006-01-02 15:04:05")
}

// GetUserName returns the username of the user who made the modification
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
	return fmt.Sprintf("Modified by %s at %s",
		m.GetUserName(),
		m.GetTime().Format("2006-01-02 15:04:05"))
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
