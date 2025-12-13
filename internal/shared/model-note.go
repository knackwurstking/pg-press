package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

const (
	// NOTE: Because i have an existing database with these levels, do not change the order or values
	LevelInfo NoteLevel = iota
	LevelAttention
	LevelBroken
	LevelNormal
)

type NoteLevel int

type Note struct {
	ID        EntityID  `json:"id"`
	Level     NoteLevel `json:"level"`
	Content   string    `json:"content"`
	CreatedAt UnixMilli `json:"created_at"`
	Linked    string    `json:"linked,omitempty"` // Generic linked entity (e.g., "press_5", "tool_123")
}

// Validate checks if the note has valid data
func (n *Note) Validate() *errors.ValidationError {
	if n.Content == "" {
		return errors.NewValidationError("note content is required")
	}

	return nil
}

// Clone creates a copy of the note
func (n *Note) Clone() *Note {
	return &Note{
		ID:        n.ID,
		Level:     n.Level,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		Linked:    n.Linked,
	}
}

// String returns a string representation of the note
func (n *Note) String() string {
	return fmt.Sprintf(
		"Note[ID=%s, Level=%d, Content=%s, CreatedAt=%s, Linked=%s]",
		n.ID.String(),
		n.Level,
		n.Content,
		n.CreatedAt.FormatDateTime(),
		n.Linked,
	)
}

// TODO: Get linked method

var _ Entity[*Note] = (*Note)(nil)
