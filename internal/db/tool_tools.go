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

const SQLAddTool string = `
	INSERT INTO tools (width, height, position, type, code, cycles_offset, cycles, is_dead, cassette, min_thickness, max_thickness)
	VALUES (:width, :height, :position, :type, :code, :cycles_offset, :cycles, :is_dead, :cassette, :min_thickness, :max_thickness);
`

func AddTool(tool *shared.Tool) *errors.MasterError {
	if verr := tool.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid tool data")
	}

	_, err := DBTool.Exec(SQLAddTool,
		sql.Named("width", tool.Width),
		sql.Named("height", tool.Height),
		sql.Named("position", tool.Position),
		sql.Named("type", tool.Type),
		sql.Named("code", tool.Code),
		sql.Named("cycles_offset", tool.CyclesOffset),
		sql.Named("cycles", tool.Cycles),
		sql.Named("is_dead", tool.IsDead),
		sql.Named("cassette", tool.Cassette),
		sql.Named("min_thickness", tool.MinThickness),
		sql.Named("max_thickness", tool.MaxThickness),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

const SQLUpdateTool string = `
	UPDATE tools
	SET width 			= :width,
		height 			= :height,
		position 		= :position,
		type 			= :type,
		code 			= :code,
		cycles_offset 	= :cycles_offset,
		cycles 			= :cycles,
		is_dead 		= :is_dead,
		cassette 		= :cassette,
		min_thickness 	= :min_thickness,
		max_thickness 	= :max_thickness
	WHERE id = :id;
`

func UpdateTool(tool *shared.Tool) *errors.MasterError {
	if verr := tool.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid tool data")
	}

	_, err := DBTool.Exec(SQLUpdateTool,
		sql.Named("id", tool.ID),
		sql.Named("width", tool.Width),
		sql.Named("height", tool.Height),
		sql.Named("position", tool.Position),
		sql.Named("type", tool.Type),
		sql.Named("code", tool.Code),
		sql.Named("cycles_offset", tool.CyclesOffset),
		sql.Named("cycles", tool.Cycles),
		sql.Named("is_dead", tool.IsDead),
		sql.Named("cassette", tool.Cassette),
		sql.Named("min_thickness", tool.MinThickness),
		sql.Named("max_thickness", tool.MaxThickness),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

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
		return nil, errors.NewMasterError(err)
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
		return errors.NewMasterError(err)
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
		return errors.NewMasterError(err)
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
		return errors.NewMasterError(err)
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
		return nil, errors.NewMasterError(err)
	}
	return &t, nil
}
