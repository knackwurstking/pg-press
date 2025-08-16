package database

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createNotesTableQuery = `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllNotesQuery = `
		SELECT id, level, content FROM notes;
	`

	selectNoteByIDQuery = `
		SELECT id, format, type, code, notes FROM notes WHERE id = $1;
	`
)

type Notes struct {
	db *sql.DB
}

func NewNotes(db *sql.DB) *Notes {
	if _, err := db.Exec(createNotesTableQuery); err != nil {
		panic(
			NewDatabaseError(
				"create_table",
				"notes",
				"failed to create notes table",
				err,
			),
		)
	}

	return &Notes{
		db: db,
	}
}

func (n *Notes) List() ([]*Note, error) {
	logger.Notes().Info("Listing notes")

	rows, err := n.db.Query(selectAllNotesQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "notes",
			"failed to query notes", err)
	}
	defer rows.Close()

	var notes []*Note

	for rows.Next() {
		note, err := n.scanNoteFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan note")
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "notes",
			"error iterating over rows", err)
	}

	return notes, nil
}

func (n *Notes) Get(id int64) (*Note, error) {
	logger.Notes().Info("Getting note, id: %d", id)

	row := n.db.QueryRow(selectNoteByIDQuery, id)

	note, err := n.scanNoteFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "notes",
			fmt.Sprintf("failed to get note with ID %d", id), err)
	}

	return note, nil
}

func (n *Notes) GetByIDs(ids []int64) ([]*Note, error) {
	if len(ids) == 0 {
		return []*Note{}, nil
	}

	logger.Notes().Debug("Getting notes by IDs: %v", ids)

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(selectAttachmentsByIDsQuery,
		joinStrings(placeholders, ","))

	rows, err := n.db.Query(query, args...)
	if err != nil {
		return nil, NewDatabaseError("select", "notes",
			"failed to query notes by IDs", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	noteMap := make(map[int64]*Note)

	for rows.Next() {
		note, err := n.scanNoteFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan attachment")
		}
		noteMap[note.ID] = note
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "notes",
			"error iterating over rows", err)
	}

	// Return attachments in the order of the requested IDs
	var notes []*Note
	for _, id := range ids {
		if note, exists := noteMap[id]; exists {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

func (n *Notes) scanNoteFromRows(rows *sql.Rows) (*Note, error) {
	note := &Note{}

	if err := rows.Scan(&note.ID, &note.Level, &note.Content); err != nil {
		return nil, NewDatabaseError("scan", "notes",
			"failed to scan row", err)
	}

	return note, nil
}

func (n *Notes) scanNoteFromRow(row *sql.Row) (*Note, error) {
	note := &Note{}

	if err := row.Scan(&note.ID, &note.Level, &note.Content); err != nil {
		return nil, NewDatabaseError("scan", "notes",
			"failed to scan row", err)
	}

	return note, nil
}
