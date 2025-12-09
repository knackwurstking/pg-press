package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/errors"
)

const (
	LevelInfo Level = iota
	LevelAttention
	LevelBroken
)

type NoteID int64

type Level int

type Linked struct {
	Name string
	ID   int64
}

type Note struct {
	ID        NoteID    `json:"id"`
	Level     Level     `json:"level"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Linked    string    `json:"linked,omitempty"` // Generic linked entity (e.g., "press_5", "tool_123")
}

func NewNote(l Level, message string) *Note {
	return &Note{
		Level:     l,
		Content:   message,
		CreatedAt: time.Now(),
		Linked:    "",
	}
}

// NewLinkedNote creates a new linked note to a specific entity
func NewLinkedNote(l Level, message string, linked string) *Note {
	return &Note{
		Level:     l,
		Content:   message,
		CreatedAt: time.Now(),
		Linked:    linked,
	}
}

func (n *Note) Validate() *errors.ValidationError {
	if n.Content == "" {
		return errors.NewValidationError("content cannot be empty")
	}

	if n.Level < 0 {
		return errors.NewValidationError("invalid level: %d", n.Level)
	}

	return nil
}

// IsLinked returns true if the note is linked to any entity
func (n *Note) IsLinked() bool {
	return n.Linked != ""
}

// GetLinked returns the linked entity string
func (n *Note) GetLinked() Linked {
	s := strings.Split(n.Linked, "_")
	id, _ := strconv.ParseInt(s[1], 10, 64)
	return Linked{Name: s[0], ID: id}
}
