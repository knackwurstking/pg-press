package database

import (
	"database/sql"

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
)

type Notes struct {
	db    *sql.DB
	feeds *Feeds
}

func NewNotes(db *sql.DB, feeds *Feeds) *Notes {
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
		db:    db,
		feeds: feeds,
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
		tool, err := n.scanNoteFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan note")
		}
		notes = append(notes, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "notes",
			"error iterating over rows", err)
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
