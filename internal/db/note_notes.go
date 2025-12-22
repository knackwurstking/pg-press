package db

import (
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
