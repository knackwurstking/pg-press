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

func ListNotesForLinked(linked string, id shared.EntityID) ([]*shared.Note, *errors.MasterError) {
	rows, err := Note.Query(SQLListNotes)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
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
		return nil, errors.NewMasterError(err, 0)
	}

	return n, nil
}
