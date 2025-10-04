package models

import "time"

const (
	INFO Level = iota
	ATTENTION
	BROKEN
)

type Level int

type Note struct {
	ID        int64     `json:"id"`
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

// NewNoteLinked creates a new note linked to a specific entity
func NewNoteLinked(l Level, message string, linked string) *Note {
	return &Note{
		Level:     l,
		Content:   message,
		CreatedAt: time.Now(),
		Linked:    linked,
	}
}

// Important returns true if the note level is ATTENTION or BROKEN
func (n *Note) IsImportant() bool {
	return n.Level == ATTENTION || n.Level == BROKEN
}

func (n *Note) IsInfo() bool {
	return n.Level == INFO
}

func (n *Note) IsAttention() bool {
	return n.Level == ATTENTION
}

func (n *Note) IsBroken() bool {
	return n.Level == BROKEN
}

// IsLinked returns true if the note is linked to any entity
func (n *Note) IsLinked() bool {
	return n.Linked != ""
}

// GetLinked returns the linked entity string
func (n *Note) GetLinked() string {
	return n.Linked
}
