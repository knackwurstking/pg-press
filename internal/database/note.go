package database

const (
	INFO NoteLevel = iota
	ATTENTION
	BROKEN
)

type NoteLevel int

type Note struct {
	ID      int64     `json:"id"`
	Level   NoteLevel `json:"level"`
	Content string    `json:"content"`
}
