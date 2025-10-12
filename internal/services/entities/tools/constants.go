package tools

const (
	ToolQueryInsert = `position, format, type, code, regenerating, is_dead, press, binding`
	ToolQuerySelect = "id, position, format, type, code, regenerating, is_dead, press, binding"
	ToolQueryUpdate = `
		position = ?,
		format = ?,
		type = ?,
		code = ?,
		regenerating = ?,
		is_dead = ?,
		press = ?,
		binding = ?
	`
)
