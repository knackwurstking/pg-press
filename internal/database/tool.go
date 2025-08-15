package database

import "fmt"

type ToolFormat struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (tf ToolFormat) String() string {
	return fmt.Sprintf("%dx%d", tf.Width, tf.Height)
}

// Tool represents a tool in the database.
type Tool struct {
	ID     int64      `json:"id"`
	Format ToolFormat `json:"format"`
	Type   string     `json:"type"` // Ex: FC, GTC, MASS
	Code   string     `json:"code"`
	Notes  []Note     `json:"notes"`
}

func (t *Tool) String() string {
	return fmt.Sprintf("%s %s (%s)", t.Format, t.Code, t.Type)
}
