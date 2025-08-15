package database

const (
	INFO NoteLevel = iota
	ATTENTION
	BROKEN
)

type NoteLevel int

type Note struct {
	Level   NoteLevel `json:"level"`
	Content string    `json:"content"`
}
