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
)

// -----------------------------------------------------------------------------
// Table Helpers: "tool_regenerations"
// -----------------------------------------------------------------------------

const SQLGetToolRegeneration string = `
	SELECT id, tool_id, start, stop, cycles
	FROM tool_regenerations
	WHERE id = :id;
`

func GetToolRegeneration(id shared.EntityID) (*shared.ToolRegeneration, *errors.MasterError) {
	row := DBUser.QueryRow(SQLGetToolRegeneration, sql.Named("id", int64(id)))
	tr, merr := ScanToolRegeneration(row)
	if merr != nil {
		return nil, merr
	}
	return tr, nil
}

const SQLListToolRegenerations string = `
	SELECT id, tool_id, start, stop, cycles
	FROM tool_regenerations
	WHERE tool_id = :tool_id;
`

func ListToolRegenerations(toolID shared.EntityID) ([]*shared.ToolRegeneration, *errors.MasterError) {
	rows, err := DBUser.Query(SQLListToolRegenerations, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
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

const SQLDeleteToolRegeneration string = `
	DELETE FROM tool_regenerations
	WHERE id = :id;
`

func DeleteToolRegeneration(id shared.EntityID) *errors.MasterError {
	_, err := DBTool.Exec(SQLDeleteToolRegeneration, sql.Named("id", int64(id)))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

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
