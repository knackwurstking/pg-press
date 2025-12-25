package db

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

const (
	sqlCreateToolsTable string = `
		CREATE TABLE IF NOT EXISTS tools (
			id					INTEGER NOT NULL, 			-- Base Tool
			width 				INTEGER NOT NULL, 			-- Base Tool
			height 				INTEGER NOT NULL, 			-- Base Tool
			position 			INTEGER NOT NULL, 			-- Base Tool
			type 				TEXT NOT NULL, 				-- Base Tool
			code 				TEXT NOT NULL, 				-- Base Tool
			cycles_offset 		INTEGER NOT NULL DEFAULT 0, -- Base Tool
			is_dead 			INTEGER NOT NULL DEFAULT 0, -- Base Tool
			cassette			INTEGER NOT NULL DEFAULT 0, -- Tool
			min_thickness		REAL NOT NULL DEFAULT 0, 	-- Cassette
			max_thickness		REAL NOT NULL DEFAULT 0, 	-- Cassette

			PRIMARY KEY("id" AUTOINCREMENT),

			-- FOREIGN KEY for cassette to tools.id
			FOREIGN KEY(cassette) REFERENCES tools(id)
		);
	`

	sqlAddTool string = `
		INSERT INTO tools (width, height, position, type, code, cycles_offset, is_dead, cassette, min_thickness, max_thickness)
		VALUES (:width, :height, :position, :type, :code, :cycles_offset, :is_dead, :cassette, :min_thickness, :max_thickness);
	`

	sqlUpdateTool string = `
		UPDATE tools
		SET width 			= :width,
			height 			= :height,
			position 		= :position,
			type 			= :type,
			code 			= :code,
			cycles_offset 	= :cycles_offset,
			is_dead 		= :is_dead,
			cassette 		= :cassette,
			min_thickness 	= :min_thickness,
			max_thickness 	= :max_thickness
		WHERE id = :id;
	`

	sqlGetTool string = `
		SELECT id, width, height, position, type, codee, cycles_offset, is_dead, cassette, min_thickness, max_thickness
		FROM tools
		WHERE id = :id;
	`

	sqlListTools string = `
		SELECT id, width, height, position, type, codee, cycles_offset, is_dead, cassette, min_thickness, max_thickness
		FROM tools
		ORDER BY id ASC;
	`

	sqlDeleteTool string = `
		DELETE FROM tools
		WHERE id = :id;
	`

	sqlMarkToolAsDead string = `
		UPDATE tools
		SET is_dead = 1
		WHERE id = :id;
	`

	sqlReviveTool string = `
		UPDATE tools
		SET is_dead = 0
		WHERE id = :id;
	`

	sqlBindTool string = `
		UPDATE tools
		SET cassette = :target_id
		WHERE id = :source_id AND cassette = 0;
	`

	sqlUnbindTool string = `
		UPDATE tools
		SET cassette = 0
		WHERE id = :source_id;
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "tools"
// -----------------------------------------------------------------------------

func AddTool(tool *shared.Tool) *errors.MasterError {
	if verr := tool.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid tool data")
	}

	_, err := dbTool.Exec(sqlAddTool,
		sql.Named("width", tool.Width),
		sql.Named("height", tool.Height),
		sql.Named("position", tool.Position),
		sql.Named("type", tool.Type),
		sql.Named("code", tool.Code),
		sql.Named("cycles_offset", tool.CyclesOffset),
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

func UpdateTool(tool *shared.Tool) *errors.MasterError {
	if verr := tool.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid tool data")
	}

	_, err := dbTool.Exec(sqlUpdateTool,
		sql.Named("id", tool.ID),
		sql.Named("width", tool.Width),
		sql.Named("height", tool.Height),
		sql.Named("position", tool.Position),
		sql.Named("type", tool.Type),
		sql.Named("code", tool.Code),
		sql.Named("cycles_offset", tool.CyclesOffset),
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

func GetTool(id shared.EntityID) (*shared.Tool, *errors.MasterError) {
	tool, merr := ScanTool(dbTool.QueryRow(sqlGetTool, id))
	if merr != nil {
		return tool, merr
	}

	merr = InjectCyclesIntoTool(tool)
	if merr != nil {
		return nil, merr
	}

	return tool, nil
}

func ListTools() (tools []*shared.Tool, merr *errors.MasterError) {
	r, err := dbTool.Query(sqlListTools)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}

	for r.Next() {
		tool, merr := ScanTool(r)
		if merr != nil {
			r.Close()
			return nil, merr
		}
		tools = append(tools, tool)
	}
	r.Close()

	for _, tool := range tools {
		merr = InjectCyclesIntoTool(tool)
		if merr != nil {
			return nil, merr
		}
	}

	return tools, nil
}

func ListBindableCassettes(id shared.EntityID) (cassettes []*shared.Tool, merr *errors.MasterError) {
	tool, merr := GetTool(id)
	if merr != nil {
		return nil, merr
	}
	if tool.IsCassette() {
		return nil, errors.NewValidationError("cannot bind cassette to itself").MasterError()
	}

	tools, merr := ListTools()
	if merr != nil {
		return nil, merr
	}
	for _, t := range tools {
		if !t.IsCassette() || t.IsDead || t.Width != tool.Width || t.Height != tool.Height {
			continue
		}
		cassettes = append(cassettes, t)
	}
	return cassettes, nil
}

func DeleteTool(id shared.EntityID) *errors.MasterError {
	_, err := dbTool.Exec(sqlDeleteTool, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func MarkToolAsDead(id shared.EntityID) *errors.MasterError {
	_, err := dbTool.Exec(sqlMarkToolAsDead, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func ReviveTool(id shared.EntityID) *errors.MasterError {
	_, err := dbTool.Exec(sqlReviveTool, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func BindTool(sourceID, targetID shared.EntityID) *errors.MasterError {
	res, err := dbTool.Exec(sqlBindTool,
		sql.Named("source_id", sourceID),
		sql.Named("target_id", targetID),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.NewMasterError(err)
	}
	if rowsAffected == 0 {
		return errors.NewMasterError(
			fmt.Errorf("tool %d is already bound to a cassette", sourceID),
		)
	}
	return nil
}

func UnbindTool(sourceID shared.EntityID) *errors.MasterError {
	_, err := dbTool.Exec(sqlUnbindTool,
		sql.Named("source_id", sourceID),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func InjectCyclesIntoTool(tool *shared.Tool) *errors.MasterError {
	cycles, merr := GetTotalToolCycles(tool.ID)
	if merr != nil {
		return merr.Wrap("could not get total cycles for tool ID %d", tool.ID)
	}
	tool.Cycles = cycles
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
