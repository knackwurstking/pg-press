package mod

import (
	"fmt"
	"slices"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models/user"
)

type Mods[T any] []*Mod[T]

func NewMods[T any](mod ...*Mod[T]) Mods[T] {
	m := Mods[T]{}
	m = append(m, mod...)
	return m
}

func (m *Mods[T]) Reversed() []*Mod[T] {
	reversed := make([]*Mod[T], len(*m))
	copy(reversed, *m)
	slices.Reverse(reversed)
	return reversed
}

func (m *Mods[T]) Add(user *user.User, data T) {
	*m = append(*m, NewMod(user, data))
}

func (m *Mods[T]) First() *Mod[T] {
	return (*m)[0]
}

func (m *Mods[T]) Current() *Mod[T] {
	if len(*m) == 0 {
		return nil
	}
	return (*m)[len(*m)-1]
}

func (m *Mods[T]) Get(time int64) (*Mod[T], error) {
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

// Mod represents a modification record that tracks changes made to any type T
type Mod[T any] struct {
	User *user.User `json:"user"`
	Time int64      `json:"time"` // Time of modification in milliseconds since Unix epoch,
	// should be unique
	Data T `json:"data"`
}

// NewMod creates a new modification record with the current timestamp
func NewMod[T any](u *user.User, data T) *Mod[T] {
	if u == nil {
		u = &user.User{Name: "system"}
	}

	return &Mod[T]{
		User: u,
		Time: time.Now().UnixMilli(),
		Data: data,
	}
}

// GetTime returns the modification time as a Go time.Time
func (m *Mod[T]) GetTime() time.Time {
	return time.UnixMilli(m.Time)
}

func (m *Mod[T]) GetTimeString() string {
	return m.GetTime().Format("2006-01-02 15:04:05")
}

// GetUserName returns the username of the user who made the modification
func (m *Mod[T]) GetUserName() string {
	if m.User == nil {
		return "unknown"
	}
	return m.User.Name
}

// IsModifiedBy checks if the modification was made by a specific user
func (m *Mod[T]) IsModifiedBy(userName string) bool {
	if m.User == nil {
		return userName == "unknown" || userName == ""
	}
	return m.User.Name == userName
}

// String returns a human-readable representation of the modification
func (m *Mod[T]) String() string {
	return fmt.Sprintf("Modified by %s at %s",
		m.GetUserName(),
		m.GetTime().Format("2006-01-02 15:04:05"))
}

// Age returns the duration since the modification was made
func (m *Mod[T]) Age() time.Duration {
	return time.Since(m.GetTime())
}

// IsOlderThan checks if the modification is older than the specified duration
func (m *Mod[T]) IsOlderThan(duration time.Duration) bool {
	return m.Age() > duration
}

// IsNewerThan checks if the modification is newer than the specified duration
func (m *Mod[T]) IsNewerThan(duration time.Duration) bool {
	return m.Age() < duration
}

// Validate checks if the modification record is valid
func (m *Mod[T]) Validate() error {
	if m.Time <= 0 {
		return fmt.Errorf("modification time must be positive")
	}
	if m.Time > time.Now().UnixMilli() {
		return fmt.Errorf("modification time cannot be in the future")
	}
	return nil
}
