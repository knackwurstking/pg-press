package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// ------------------------------------------------------------------------------

const (
	sqlCreateToolRegenerationsTable string = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id 		INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			start 	INTEGER NOT NULL,
			stop 	INTEGER NOT NULL DEFAULT 0,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	sqlAddToolRegeneration string = `
		INSERT INTO tool_regenerations (tool_id, start, stop)
		VALUES (:tool_id, :start, :stop);
	`

	sqlUpdateToolRegeneration string = `
		UPDATE tool_regenerations
		SET
			tool_id = :tool_id,
			start = :start,
			stop = :stop,
		WHERE id = :id;
	`

	sqlGetToolRegeneration string = `
		SELECT id, tool_id, start, stop
		FROM tool_regenerations
		WHERE id = :id;
	`

	sqlListToolRegenerations string = `
		SELECT id, tool_id, start, stop
		FROM tool_regenerations
		ORDER BY start DESC
	;
	`

	sqlListToolRegenerationsByTool string = `
		SELECT id, tool_id, start, stop
		FROM tool_regenerations
		WHERE tool_id = :tool_id;
	`

	sqlDeleteToolRegeneration string = `
		DELETE FROM tool_regenerations
		WHERE id = :id;
	`

	sqlDeleteToolRegenerationByTool string = `
		DELETE FROM tool_regenerations
		WHERE tool_id = :tool_id;
	`

	sqlToolRegenerationInProgress string = `
		SELECT COUNT(*)
		FROM tool_regenerations
		WHERE stop = 0 AND tool_id = :tool_id
		ORDER BY start DESC;
	`

	sqlStartToolRegeneration string = `
		INSERT INTO tool_regenerations (tool_id, start)
		VALUES (:tool_id, :start);
	`

	sqlStopToolRegeneration string = `
		UPDATE tool_regenerations
		SET stop = :stop
		WHERE tool_id = :tool_id AND stop = 0;
	`
)

// -----------------------------------------------------------------------------
// Tool Regeneration Functions
// -----------------------------------------------------------------------------

// AddToolRegeneration adds a new tool regeneration to the database
func AddToolRegeneration(tr *shared.ToolRegeneration) *errors.MasterError {
	if verr := tr.Validate(); verr != nil {
		return errors.NewMasterError(verr)
	}

	_, err := dbTool.Exec(sqlAddToolRegeneration,
		sql.Named("tool_id", tr.ToolID),
		sql.Named("start", tr.Start),
		sql.Named("stop", tr.Stop),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// UpdateToolRegeneration updates an existing tool regeneration in the database
func UpdateToolRegeneration(tr *shared.ToolRegeneration) *errors.MasterError {
	if verr := tr.Validate(); verr != nil {
		return errors.NewMasterError(verr)
	}

	_, err := dbTool.Exec(sqlUpdateToolRegeneration,
		sql.Named("id", tr.ID),
		sql.Named("tool_id", tr.ToolID),
		sql.Named("start", tr.Start),
		sql.Named("stop", tr.Stop),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// GetToolRegeneration retrieves a tool regeneration by its ID
func GetToolRegeneration(id shared.EntityID) (*shared.ToolRegeneration, *errors.MasterError) {
	row := dbTool.QueryRow(sqlGetToolRegeneration, sql.Named("id", int64(id)))
	tr, merr := ScanToolRegeneration(row)
	if merr != nil {
		return nil, merr
	}
	return tr, nil
}

// ListToolRegenerations retrieves all tool regenerations from the database
func ListToolRegenerations() ([]*shared.ToolRegeneration, *errors.MasterError) {
	r, err := dbTool.Query(sqlListToolRegenerations)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer r.Close()

	var trs []*shared.ToolRegeneration
	for r.Next() {
		tr, merr := ScanToolRegeneration(r)
		if merr != nil {
			return nil, merr
		}
		trs = append(trs, tr)
	}

	return trs, nil
}

// ListToolRegenerationsByTool retrieves all tool regenerations for a specific tool
func ListToolRegenerationsByTool(toolID shared.EntityID) ([]*shared.ToolRegeneration, *errors.MasterError) {
	rows, err := dbTool.Query(sqlListToolRegenerationsByTool, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewMasterError(err)
	}

	var trs []*shared.ToolRegeneration
	for rows.Next() {
		tr, merr := ScanToolRegeneration(rows)
		if merr != nil {
			rows.Close()
			return nil, merr
		}
		trs = append(trs, tr)
	}
	rows.Close()

	return trs, nil
}

// DeleteToolRegeneration removes a tool regeneration from the database
func DeleteToolRegeneration(id shared.EntityID) *errors.MasterError {
	_, err := dbTool.Exec(sqlDeleteToolRegeneration, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// DeleteToolRegenerationByTool removes all tool regenerations for a specific tool
func DeleteToolRegenerationByTool(toolID shared.EntityID) *errors.MasterError {
	_, err := dbTool.Exec(sqlDeleteToolRegenerationByTool, sql.Named("tool_id", toolID))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// ToolRegenerationInProgress checks if a tool regeneration is currently in progress
func ToolRegenerationInProgress(toolID shared.EntityID) (bool, *errors.MasterError) {
	row := dbTool.QueryRow(sqlToolRegenerationInProgress, sql.Named("tool_id", toolID))
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, errors.NewMasterError(err)
	}
	return count > 0, nil
}

// StartToolRegeneration starts a new tool regeneration
func StartToolRegeneration(toolID shared.EntityID) *errors.MasterError {
	tool, merr := GetTool(toolID)
	if merr != nil {
		return merr.Wrap("getting tool by ID failed")
	}

	// Check if a already started regeneration exists for this tool
	if inProgress, merr := ToolRegenerationInProgress(tool.ID); merr != nil {
		return merr.Wrap("checking for in-progress regeneration failed (Tool ID %d)", tool.ID)
	} else if inProgress {
		return errors.NewMasterError(fmt.Errorf("a tool regeneration is already in progress for tool with ID %d", tool.ID))
	}

	_, err := dbTool.Exec(sqlStartToolRegeneration,
		sql.Named("tool_id", tool.ID),
		sql.Named("start", shared.NewUnixMilli(time.Now())),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}

// StopToolRegeneration stops an ongoing tool regeneration
func StopToolRegeneration(toolID shared.EntityID) *errors.MasterError {
	tool, merr := GetTool(toolID)
	if merr != nil {
		return merr.Wrap("getting tool by ID failed")
	}

	_, err := dbTool.Exec(sqlStopToolRegeneration,
		sql.Named("tool_id", toolID),
		sql.Named("stop", shared.NewUnixMilli(time.Now())),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}

	// Reset tool cycles to zero
	tool.CyclesOffset = 0
	tool.Cycles = 0
	merr = UpdateTool(tool)
	if merr != nil {
		return merr.Wrap("updating tool after regeneration failed")
	}

	return nil
}

// AbortToolRegeneration aborts an ongoing tool regeneration
func AbortToolRegeneration(toolID shared.EntityID) *errors.MasterError {
	merr := DeleteToolRegenerationByTool(toolID)
	if merr != nil {
		return merr.Wrap("deleting tool regeneration failed")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanToolRegeneration scans a database row into a ToolRegeneration struct
func ScanToolRegeneration(row Scannable) (*shared.ToolRegeneration, *errors.MasterError) {
	var tr shared.ToolRegeneration
	err := row.Scan(
		&tr.ID,
		&tr.ToolID,
		&tr.Start,
		&tr.Stop,
	)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	return &tr, nil
}
