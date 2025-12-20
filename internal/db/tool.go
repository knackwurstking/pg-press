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
	SQLCreateMetalSheetsTable string = `
		CREATE TABLE IF NOT EXISTS metal_sheets (
			id 				INTEGER NOT NULL,
			tool_id 		INTEGER NOT NULL,
			tile_height 	REAL NOT NULL,
			value 			REAL NOT NULL,
			type 			TEXT NOT NULL,
			marke_height 	INTEGER,
			stf 			REAL,
			stf_max 		REAL,
			identifier 		TEXT,

			PRIMARY KEY("id" AUTOINCREMENT),
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
	`

	SQLCreateToolRegenerationsTable string = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id 		INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			start 	INTEGER NOT NULL,
			stop 	INTEGER NOT NULL,
			cycles 	INTEGER NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

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

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanUpperMetalSheet(row Scannable) (*shared.UpperMetalSheet, *errors.MasterError) {
	var ums shared.UpperMetalSheet
	err := row.Scan(
		&ums.ID,
		&ums.ToolID,
		&ums.TileHeight,
		&ums.Value,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &ums, nil
}

func ScanLowerMetalSheet(row Scannable) (*shared.LowerMetalSheet, *errors.MasterError) {
	var lms shared.LowerMetalSheet
	err := row.Scan(
		&lms.ID,
		&lms.ToolID,
		&lms.TileHeight,
		&lms.Value,
		&lms.MarkeHeight,
		&lms.STF,
		&lms.STFMax,
		&lms.Identifier,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &lms, nil
}

func ScanToolRegeneration(row Scannable) (*shared.ToolRegeneration, *errors.MasterError) {
	var tr shared.ToolRegeneration
	err := row.Scan(
		&tr.ID,
		&tr.ToolID,
		&tr.Start,
		&tr.Stop,
		&tr.Cycles,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &tr, nil
}

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
