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
	SQLCreateNotesTable string = `
		CREATE TABLE IF NOT EXISTS notes (
			id 			INTEGER NOT NULL,
			level 		INTEGER NOT NULL,
			content 	TEXT NOT NULL,
			created_at 	INTEGER NOT NULL,
			linked 		TEXT NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
)

// -----------------------------------------------------------------------------
// SQL Queries
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// Table Helpers: "notes"
// -----------------------------------------------------------------------------

const SQLAddNote string = `
	INSERT INTO notes (level, content, created_at, linked)
	VALUES (:level, :content, :created_at, :linked);
`

func AddNote(note *shared.Note) *errors.MasterError {
	if verr := note.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid note data")
	}

	_, err := DBNote.Exec(SQLAddNote,
		sql.Named("level", note.Level),
		sql.Named("content", note.Content),
		sql.Named("created_at", note.CreatedAt),
		sql.Named("linked", note.Linked),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

const SQLUpdateNote string = `
	UPDATE notes
	SET level 		= :level,
		content 	= :content,
		created_at 	= :created_at,
		linked 		= :linked
	WHERE id = :id;
`

func UpdateNote(note *shared.Note) *errors.MasterError {
	if verr := note.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid note data")
	}

	_, err := DBNote.Exec(SQLUpdateNote,
		sql.Named("id", note.ID),
		sql.Named("level", note.Level),
		sql.Named("content", note.Content),
		sql.Named("created_at", note.CreatedAt),
		sql.Named("linked", note.Linked),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

const SQLGetNote string = `
	SELECT id, level, content, created_at, linked
	FROM notes
	WHERE id = ?;
`

func GetNote(id shared.EntityID) (*shared.Note, *errors.MasterError) {
	row := DBNote.QueryRow(SQLGetNote, id)
	note, merr := ScanNote(row)
	if merr != nil {
		return nil, merr
	}
	return note, nil
}

const SQLListNotes string = `
	SELECT id, level, content, created_at, linked
	FROM notes
	ORDER BY created_at DESC;
`

func ListNotes() ([]*shared.Note, *errors.MasterError) {
	rows, err := DBNote.Query(SQLListNotes)
	if err != nil {
		return nil, errors.NewMasterError(err)
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

const SQLListNotesForLinked string = `
	SELECT id, level, content, created_at, linked
	FROM notes
	ORDER BY created_at DESC;
`

func ListNotesForLinked(linked string, id shared.EntityID) ([]*shared.Note, *errors.MasterError) {
	rows, err := DBNote.Query(SQLListNotesForLinked)
	if err != nil {
		return nil, errors.NewMasterError(err)
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

const SQLDeleteNote string = `
	DELETE FROM notes
	WHERE id = :id;
`

func DeleteNote(id shared.EntityID) *errors.MasterError {
	_, err := DBNote.Exec(SQLDeleteNote, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanNote(row Scannable) (*shared.Note, *errors.MasterError) {
	n := &shared.Note{}
	err := row.Scan(
		&n.ID,
		&n.Level,
		&n.Content,
		&n.CreatedAt,
		&n.Linked,
	)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}

	return n, nil
}
