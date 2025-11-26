package services

import (
	"fmt"
)

const TableNameTools = "tools"

const (
	ToolQueryInsert = `position, format, type, code, regenerating, is_dead, press, binding`
	ToolQuerySelect = `id, position, format, type, code, regenerating, is_dead, press, binding`
	ToolQueryUpdate = `position = ?, format = ?, type = ?, code = ?, regenerating = ?, is_dead = ?, press = ?, binding = ?`
)

type Tools struct {
	*Base
}

func NewTools(r *Registry) *Tools {
	t := &Tools{
		Base: NewBase(r),
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			regenerating INTEGER NOT NULL DEFAULT 0,
			is_dead INTEGER NOT NULL DEFAULT 0,
			press INTEGER,
			binding INTEGER,
			PRIMARY KEY("id" AUTOINCREMENT)
		)`, TableNameTools)

	if err := t.CreateTable(query, TableNameTools); err != nil {
		panic(err)
	}

	return t
}
