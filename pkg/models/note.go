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
}

func NewNote(l Level, message string) *Note {
	return &Note{
		Level:     l,
		Content:   message,
		CreatedAt: time.Now(),
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
