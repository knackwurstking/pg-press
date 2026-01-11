package shared

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/errors"
)

const (
	// NOTE: Because i have an existing database with these levels, do not change the order or values
	LevelNormal NoteLevel = iota
	LevelInfo
	LevelAttention
	LevelBroken
)

type NoteLevel int

func (nl NoteLevel) IsValid() bool {
	return nl >= LevelNormal && nl <= LevelBroken
}

type Note struct {
	ID        EntityID  `json:"id"`
	Level     NoteLevel `json:"level"`
	Content   string    `json:"content"`
	CreatedAt UnixMilli `json:"created_at"`
	Linked    string    `json:"linked,omitempty"` // Generic linked entity (e.g., "press_5", "tool_123")
}

// Validate checks if the note has valid data
func (n *Note) Validate() *errors.ValidationError {
	if n.CreatedAt == 0 {
		return errors.NewValidationError("note creation timestamp is required")
	}
	if !n.Level.IsValid() {
		return errors.NewValidationError("note level must be one of: 0 (Normal), 1 (Info), 2 (Attention), or 3 (Broken)")
	}
	if n.Content == "" {
		return errors.NewValidationError("note content cannot be empty")
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
		"Note{ID:%s, Level:%d, Content:%s, CreatedAt:%s, Linked:%s}",
		n.ID.String(),
		n.Level,
		n.Content,
		n.CreatedAt.FormatDateTime(),
		n.Linked,
	)
}

type Linked struct {
	Name string
	ID   int64
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
