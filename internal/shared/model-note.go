package shared

import "time"

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
	CreatedAt time.Time `json:"created_at"`
	Linked    string    `json:"linked,omitempty"` // Generic linked entity (e.g., "press_5", "tool_123")
}

// TODO: Implement all the missing methods for the Note struct to satisfy the Entity interface

var _ Entity[*Note] = (*Note)(nil)
