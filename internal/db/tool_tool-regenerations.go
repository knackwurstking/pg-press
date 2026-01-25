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

	sqlAddToolRegenerationWithID string = `
		INSERT INTO tool_regenerations (id, tool_id, start, stop)
		VALUES (:id, :tool_id, :start, :stop);
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
		ORDER BY stop DESC
	;
	`

	sqlListToolRegenerationsByTool string = `
		SELECT id, tool_id, start, stop
		FROM tool_regenerations
		WHERE tool_id = :tool_id
		ORDER BY stop DESC;
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
		ORDER BY stop DESC;
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
func AddToolRegeneration(tr *shared.ToolRegeneration) *errors.HTTPError {
	if err := tr.Validate(); err != nil {
		return errors.NewHTTPError(err)
	}

	var query string
	{
		if tr.ID > 0 {
			query = sqlAddToolRegenerationWithID
		} else {
			query = sqlAddToolRegeneration
		}
	}

	var queryArgs []any
	{
		if tr.ID > 0 {
			queryArgs = append(queryArgs, sql.Named("id", tr.ID))
		}
		queryArgs = append(queryArgs,
			sql.Named("tool_id", tr.ToolID),
			sql.Named("start", tr.Start),
			sql.Named("stop", tr.Stop),
		)
	}

	if _, err := dbTool.Exec(query, queryArgs...); err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// UpdateToolRegeneration updates an existing tool regeneration in the database
func UpdateToolRegeneration(tr *shared.ToolRegeneration) *errors.HTTPError {
	if err := tr.Validate(); err != nil {
		return errors.NewHTTPError(err)
	}

	_, err := dbTool.Exec(sqlUpdateToolRegeneration,
		sql.Named("id", tr.ID),
		sql.Named("tool_id", tr.ToolID),
		sql.Named("start", tr.Start),
		sql.Named("stop", tr.Stop),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetToolRegeneration retrieves a tool regeneration by its ID
func GetToolRegeneration(id shared.EntityID) (*shared.ToolRegeneration, *errors.HTTPError) {
	row := dbTool.QueryRow(sqlGetToolRegeneration, sql.Named("id", int64(id)))
	tr, herr := ScanToolRegeneration(row)
	if herr != nil {
		return nil, herr
	}
	return tr, nil
}

// ListToolRegenerations retrieves all tool regenerations from the database
func ListToolRegenerations() ([]*shared.ToolRegeneration, *errors.HTTPError) {
	r, err := dbTool.Query(sqlListToolRegenerations)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer r.Close()

	var tres []*shared.ToolRegeneration
	for r.Next() {
		tr, herr := ScanToolRegeneration(r)
		if herr != nil {
			return nil, herr
		}
		tres = append(tres, tr)
	}

	return tres, nil
}

// ListToolRegenerationsByTool retrieves all tool regenerations for a specific tool
func ListToolRegenerationsByTool(toolID shared.EntityID) ([]*shared.ToolRegeneration, *errors.HTTPError) {
	rows, err := dbTool.Query(sqlListToolRegenerationsByTool, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}

	var tres []*shared.ToolRegeneration
	for rows.Next() {
		tr, herr := ScanToolRegeneration(rows)
		if herr != nil {
			rows.Close()
			return nil, herr
		}
		tres = append(tres, tr)
	}
	rows.Close()

	return tres, nil
}

// DeleteToolRegeneration removes a tool regeneration from the database
func DeleteToolRegeneration(id shared.EntityID) *errors.HTTPError {
	_, err := dbTool.Exec(sqlDeleteToolRegeneration, sql.Named("id", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// DeleteToolRegenerationByTool removes all tool regenerations for a specific tool
func DeleteToolRegenerationByTool(toolID shared.EntityID) *errors.HTTPError {
	_, err := dbTool.Exec(sqlDeleteToolRegenerationByTool, sql.Named("tool_id", toolID))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// ToolRegenerationInProgress checks if a tool regeneration is currently in progress
func ToolRegenerationInProgress(toolID shared.EntityID) (bool, *errors.HTTPError) {
	row := dbTool.QueryRow(sqlToolRegenerationInProgress, sql.Named("tool_id", toolID))
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, errors.NewHTTPError(err)
	}
	return count > 0, nil
}

// StartToolRegeneration starts a new tool regeneration
func StartToolRegeneration(toolID shared.EntityID) *errors.HTTPError {
	tool, herr := GetTool(toolID)
	if herr != nil {
		return herr.Wrap("getting tool by ID failed")
	}

	// Check if a already started regeneration exists for this tool
	if inProgress, herr := ToolRegenerationInProgress(tool.ID); herr != nil {
		return herr.Wrap("checking for in-progress regeneration failed (Tool ID %d)", tool.ID)
	} else if inProgress {
		return errors.NewHTTPError(fmt.Errorf("a tool regeneration is already in progress for tool with ID %d", tool.ID))
	}

	_, err := dbTool.Exec(sqlStartToolRegeneration,
		sql.Named("tool_id", tool.ID),
		sql.Named("start", shared.NewUnixMilli(time.Now())),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}

	return nil
}

// StopToolRegeneration stops an ongoing tool regeneration
func StopToolRegeneration(toolID shared.EntityID) *errors.HTTPError {
	tool, herr := GetTool(toolID)
	if herr != nil {
		return herr.Wrap("getting tool by ID failed")
	}

	_, err := dbTool.Exec(sqlStopToolRegeneration,
		sql.Named("tool_id", toolID),
		sql.Named("stop", shared.NewUnixMilli(time.Now())),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}

	// Reset tool cycles to zero
	tool.CyclesOffset = 0
	tool.Cycles = 0
	herr = UpdateTool(tool)
	if herr != nil {
		return herr.Wrap("updating tool after regeneration failed")
	}

	return nil
}

// AbortToolRegeneration aborts an ongoing tool regeneration
func AbortToolRegeneration(toolID shared.EntityID) *errors.HTTPError {
	herr := DeleteToolRegenerationByTool(toolID)
	if herr != nil {
		return herr.Wrap("deleting tool regeneration failed")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanToolRegeneration scans a database row into a ToolRegeneration struct
func ScanToolRegeneration(row Scannable) (*shared.ToolRegeneration, *errors.HTTPError) {
	var tr shared.ToolRegeneration
	err := row.Scan(
		&tr.ID,
		&tr.ToolID,
		&tr.Start,
		&tr.Stop,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return &tr, nil
}
