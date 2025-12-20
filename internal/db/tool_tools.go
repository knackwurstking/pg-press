package db

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// ------------------------------------------------------------------------------

const (
	SQLCreateToolsTable string = `
		CREATE TABLE IF NOT EXISTS tools (
			id					INTEGER NOT NULL, 			-- Base Tool
			width 				INTEGER NOT NULL, 			-- Base Tool
			height 				INTEGER NOT NULL, 			-- Base Tool
			position 			INTEGER NOT NULL, 			-- Base Tool
			type 				TEXT NOT NULL, 				-- Base Tool
			code 				TEXT NOT NULL, 				-- Base Tool
			cycles_offset 		INTEGER NOT NULL DEFAULT 0, -- Base Tool
			cycles 				INTEGER NOT NULL DEFAULT 0, -- Base Tool
			is_dead 			INTEGER NOT NULL DEFAULT 0, -- Base Tool
			cassette			INTEGER NOT NULL DEFAULT 0, -- Tool
			min_thickness		REAL NOT NULL DEFAULT 0, 	-- Cassette
			max_thickness		REAL NOT NULL DEFAULT 0, 	-- Cassette

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "tools"
// -----------------------------------------------------------------------------

const SQLGetTool string = `
	SELECT id, width, height, position, type, codee, cycles_offset, cycles, is_dead, cassette, min_thickness, max_thickness
	FROM tools
	WHERE id = :id;
`

func GetTool(id shared.EntityID) (*shared.Tool, *errors.MasterError) {
	return ScanTool(DBTool.QueryRow(SQLGetTool, id))
}

const SQLListTools string = `
	SELECT id, width, height, position, type, codee, cycles_offset, cycles, is_dead, cassette, min_thickness, max_thickness
	FROM tools
	ORDER BY id ASC;
`

func ListTools() (tools []*shared.Tool, merr *errors.MasterError) {
	r, err := DBTool.Query(SQLListTools)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer r.Close()

	for r.Next() {
		tool, merr := ScanTool(r)
		if merr != nil {
			return nil, merr
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

const SQLDeleteTool string = `
	DELETE FROM tools
	WHERE id = :id;
`

func DeleteTool(id shared.EntityID) *errors.MasterError {
	_, err := DBTool.Exec(SQLDeleteTool, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

const SQLMarkToolAsDead string = `
	UPDATE tools
	SET is_dead = 1
	WHERE id = :id;
`

func MarkToolAsDead(id shared.EntityID) *errors.MasterError {
	_, err := DBTool.Exec(SQLMarkToolAsDead, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

const SQLReviveTool string = `
	UPDATE tools
	SET is_dead = 0
	WHERE id = :id;
`

func ReviveTool(id shared.EntityID) *errors.MasterError {
	_, err := DBTool.Exec(SQLReviveTool, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanTool(row Scannable) (*shared.Tool, *errors.MasterError) {
	var t shared.Tool
	err := row.Scan(
		&t.ID,
		&t.Position,
		&t.Width,
		&t.Height,
		&t.Type,
		&t.Code,
		&t.CyclesOffset,
		&t.Cycles,
		&t.IsDead,
		&t.Cassette,
		&t.MinThickness,
		&t.MaxThickness,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &t, nil
}
