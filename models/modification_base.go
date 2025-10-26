// TODO: Remove useless stuff
//
// Data:
//   - id (int64)

//   - type (string) Modification type, for example: (in short the database table names)
//	   	- "trouble_reports"
//	   	- "tool_regenerations"
//	   	- "users"
//	   	- "tools"
//	   	- "press_cycles"

// - data (any)
// - created at (time.Time)
// - user id (int64)
package models

import (
	"fmt"
	"slices"
	"time"

	"github.com/knackwurstking/pg-press/errors"
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

func (m *Mods[T]) Add(user *User, data T) {
	*m = append(*m, NewMod(user, data))
}

func (m *Mods[T]) First() *Mod[T] {
	return (*m)[0]
}

func (m *Mods[T]) Current() *Mod[T] {
	if len(*m) == 0 {
		return nil
	}

	return (*m)[0]
}

func (m *Mods[T]) Get(time int64) (*Mod[T], error) {
	for _, mod := range *m {
		if mod.Time == time {
			return mod, nil
		}
	}
	return nil, errors.NewNotFoundError("modification")
}

type Mod[T any] struct {
	User *User `json:"user"`
	Time int64 `json:"time"`
	Data T     `json:"data"`
}

func NewMod[T any](u *User, data T) *Mod[T] {
	if u == nil {
		u = &User{Name: "system"}
	}

	return &Mod[T]{
		User: u,
		Time: time.Now().UnixMilli(),
		Data: data,
	}
}

func (m *Mod[T]) GetTime() time.Time {
	return time.UnixMilli(m.Time)
}

func (m *Mod[T]) GetTimeString() string {
	return m.GetTime().Format("2006-01-02 15:04:05")
}

func (m *Mod[T]) GetUserName() string {
	if m.User == nil {
		return "unknown"
	}
	return m.User.Name
}

func (m *Mod[T]) IsModifiedBy(userName string) bool {
	if m.User == nil {
		return userName == "unknown" || userName == ""
	}
	return m.User.Name == userName
}

func (m *Mod[T]) String() string {
	return fmt.Sprintf("Modified by %s at %s",
		m.GetUserName(),
		m.GetTime().Format("2006-01-02 15:04:05"))
}

func (m *Mod[T]) Age() time.Duration {
	return time.Since(m.GetTime())
}

func (m *Mod[T]) IsOlderThan(duration time.Duration) bool {
	return m.Age() > duration
}

func (m *Mod[T]) IsNewerThan(duration time.Duration) bool {
	return m.Age() < duration
}

func (m *Mod[T]) Validate() error {
	if m.Time <= 0 {
		return fmt.Errorf("modification time must be positive")
	}
	if m.Time > time.Now().UnixMilli() {
		return fmt.Errorf("modification time cannot be in the future")
	}
	return nil
}
