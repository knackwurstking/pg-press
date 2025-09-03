package models

import "time"

const (
	INFO NoteLevel = iota
	ATTENTION
	BROKEN
)

type NoteLevel int

type Note struct {
	ID        int64     `json:"id"`
	Level     NoteLevel `json:"level"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Important returns true if the note level is ATTENTION or BROKEN
func (n *Note) Important() bool {
	return n.Level == ATTENTION || n.Level == BROKEN
}
