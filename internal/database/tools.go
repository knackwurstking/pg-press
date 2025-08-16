package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createToolsTableQuery = `
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			notes BLOB NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllToolsQuery = `
		SELECT id, format, type, code, notes FROM tools;
	`

	selectToolByIDQuery = `
		SELECT id, format, type, code, notes FROM tools WHERE id = $1;
	`

	// TODO: Implement the Tools struct.
	insertToolQuery = `
		INSERT INTO tools (format, type, code, notes) VALUES ($1, $2, $3, $4);
	`

	// TODO: Implement the Tools struct.
	updateToolQuery = `
		UPDATE tools SET format = $1, type = $2, code = $3, notes = $4 WHERE id = $5;
	`

	// TODO: Implement the Tools struct.
	deleteToolQuery = `
		DELETE FROM tools WHERE id = $1;
	`
)

// Tools represents a collection of tools in the database.
type Tools struct {
	db    *sql.DB
	feeds *Feeds
}

func NewTools(db *sql.DB, feeds *Feeds) *Tools {
	if _, err := db.Exec(createToolsTableQuery); err != nil {
		panic(
			NewDatabaseError(
				"create_table",
				"tools",
				"failed to create tools table",
				err,
			),
		)
	}

	return &Tools{
		db:    db,
		feeds: feeds,
	}
}

func (t *Tools) List() ([]*Tool, error) {
	logger.Tools().Info("Listing tools")

	rows, err := t.db.Query(selectAllToolsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "tools",
			"failed to query tools", err)
	}
	defer rows.Close()

	var tools []*Tool

	for rows.Next() {
		tool, err := t.scanToolFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan tool")
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "tools",
			"error iterating over rows", err)
	}

	return tools, nil
}

func (t *Tools) Get(id int64) (*Tool, error) {
	logger.Tools().Info("Getting tool, id: %d", id)

	row := t.db.QueryRow(selectToolByIDQuery, id)

	tool, err := t.scanToolFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to get tool with ID %d", id), err)
	}

	return tool, nil
}

func (t *Tools) scanToolFromRows(rows *sql.Rows) (*Tool, error) {
	tool := &Tool{}

	var (
		format      []byte
		linkedNotes []byte
		mods        []byte
	)

	if err := rows.Scan(&tool.ID, &format, &tool.Type,
		&tool.Code, &linkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to scan row", err)
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal format", err)
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return tool, nil
}

func (t *Tools) scanToolFromRow(row *sql.Row) (*Tool, error) {
	tool := &Tool{}

	var (
		format      []byte
		linkedNotes []byte
		mods        []byte
	)

	if err := row.Scan(&tool.ID, &format, &tool.Type,
		&tool.Code, &linkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to scan row", err)
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal format", err)
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return tool, nil
}
