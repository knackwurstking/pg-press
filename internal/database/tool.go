package database

import "fmt"

const (
	PositionTop    = "top"
	PositionBottom = "bottom"
)

type Position string

type ToolFormat struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (tf ToolFormat) String() string {
	return fmt.Sprintf("%dx%d", tf.Width, tf.Height)
}

// Tool represents a tool in the database.
type Tool struct {
	ID          int64      `json:"id"`
	Position    Position   `json:"position"`
	Format      ToolFormat `json:"format"`
	Type        string     `json:"type"`  // Ex: FC, GTC, MASS
	Code        string     `json:"code"`  // Ex: G01, G02, ...
	LinkedNotes []int64    `json:"notes"` // Contains note ids from the "notes" table
}

func (t *Tool) String() string {
	switch t.Position {
	case PositionTop:
		return fmt.Sprintf("%s %s (%s, ↑)", t.Format, t.Code, t.Type)
	case PositionBottom:
		return fmt.Sprintf("%s %s (%s, ↓)", t.Format, t.Code, t.Type)
	default:
		return fmt.Sprintf("%s %s (%s)", t.Format, t.Code, t.Type)
	}
}
