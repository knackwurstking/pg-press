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
	sqlCreateNotesTable string = `
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER NOT NULL,
	level INTEGER NOT NULL,
	content TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	linked TEXT NOT NULL,

	PRIMARY KEY("id" AUTOINCREMENT)
);`

	sqlAddNote string = `
INSERT INTO notes (level, content, created_at, linked)
VALUES (:level, :content, :created_at, :linked);`

	sqlAddNoteWithID string = `
INSERT INTO notes (id, level, content, created_at, linked)
VALUES (:id, :level, :content, :created_at, :linked);`

	sqlUpdateNote string = `
UPDATE notes
SET level = :level,
	content = :content,
	created_at = :created_at,
	linked = :linked
WHERE id = :id;`

	sqlGetNote string = `
SELECT id, level, content, created_at, linked
FROM notes
WHERE id = :id;`

	sqlListNotes string = `
SELECT id, level, content, created_at, linked
FROM notes
ORDER BY created_at DESC;`

	sqlListNotesForLinked string = `
SELECT id, level, content, created_at, linked
FROM notes
ORDER BY created_at DESC;`

	sqlDeleteNote string = `
DELETE FROM notes
WHERE id = :id;`
)

// -----------------------------------------------------------------------------
// Note Functions
// -----------------------------------------------------------------------------

// AddNote adds a new note to the database
func AddNote(note *shared.Note) *errors.HTTPError {
	if verr := note.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid note data")
	}

	var query string
	if note.ID > 0 {
		query = sqlAddNoteWithID
	} else {
		query = sqlAddNote
	}

	var queryArgs []any
	if note.ID > 0 {
		queryArgs = append(queryArgs, sql.Named("id", note.ID))
	}
	queryArgs = append(queryArgs,
		sql.Named("level", note.Level),
		sql.Named("content", note.Content),
		sql.Named("created_at", note.CreatedAt),
		sql.Named("linked", note.Linked),
	)

	if _, err := dbNote.Exec(query, queryArgs...); err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// UpdateNote updates an existing note in the database
func UpdateNote(note *shared.Note) *errors.HTTPError {
	if verr := note.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid note data")
	}

	_, err := dbNote.Exec(sqlUpdateNote,
		sql.Named("id", note.ID),
		sql.Named("level", note.Level),
		sql.Named("content", note.Content),
		sql.Named("created_at", note.CreatedAt),
		sql.Named("linked", note.Linked),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetNote retrieves a note by its ID
func GetNote(id shared.EntityID) (*shared.Note, *errors.HTTPError) {
	return ScanNote(dbNote.QueryRow(sqlGetNote, sql.Named("id", id)))
}

// ListNotes retrieves all notes from the database
func ListNotes() ([]*shared.Note, *errors.HTTPError) {
	rows, err := dbNote.Query(sqlListNotes)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer rows.Close()

	var notes []*shared.Note
	for rows.Next() {
		note, merr := ScanNote(rows)
		if merr != nil {
			return nil, merr
		}
		notes = append(notes, note)
	}
	return notes, nil
}

// ListNotesForLinked retrieves notes linked to a specific entity
func ListNotesForLinked(linked string, id int) ([]*shared.Note, *errors.HTTPError) {
	rows, err := dbNote.Query(sqlListNotesForLinked)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer rows.Close()

	var notes []*shared.Note
	for rows.Next() {
		note, merr := ScanNote(rows)
		if merr != nil {
			return nil, merr
		}
		notes = append(notes, note)
	}

	n := 0
	matchingLinked := fmt.Sprintf("%s_%d", linked, id)
	for _, note := range notes {
		if note.Linked != matchingLinked {
			continue
		}

		notes[n] = note
		n++
	}
	return notes[:n], nil
}

// DeleteNote removes a note from the database
func DeleteNote(id shared.EntityID) *errors.HTTPError {
	_, err := dbNote.Exec(sqlDeleteNote, sql.Named("id", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanNote scans a database row into a Note struct
func ScanNote(row Scannable) (*shared.Note, *errors.HTTPError) {
	n := &shared.Note{}
	err := row.Scan(
		&n.ID,
		&n.Level,
		&n.Content,
		&n.CreatedAt,
		&n.Linked,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}

	return n, nil
}
