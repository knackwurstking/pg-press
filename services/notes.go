package services

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

type Notes struct {
	*Base
}

func NewNotes(r *Registry) *Notes {
	return &Notes{
		Base: NewBase(r),
	}
}

func (n *Notes) Get(id models.NoteID) (*models.Note, *errors.MasterError) {
	row := n.DB.QueryRow(SQLGetNote, id)
	note, err := ScanNote(row)
	if err != nil {
		return note, errors.NewMasterError(err, 0)
	}
	return note, nil
}

func (n *Notes) List() ([]*models.Note, *errors.MasterError) {
	rows, err := n.DB.Query(SQLListNotes)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanNote)
}

func (n *Notes) ListByIDs(ids []models.NoteID) ([]*models.Note, *errors.MasterError) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	rows, err := n.DB.Query(
		fmt.Sprintf(SQLListNotesByIDs, strings.Join(placeholders, ",")),
		args...,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	// Store notes in a map for efficient lookup
	notes, dberr := ScanRows(rows, ScanNote)
	if dberr != nil {
		return nil, dberr
	}
	notesMap := MapNotes(notes)

	// Return notes in the order of the requested IDs
	notesOrdered := make([]*models.Note, 0, len(ids))
	for _, id := range ids {
		if n, ok := notesMap[id]; ok {
			notesOrdered = append(notesOrdered, n)
		}
	}

	return notes, nil
}

func (n *Notes) ListByLinked(name string, id int64) ([]*models.Note, *errors.MasterError) {
	rows, err := n.DB.Query(SQLListNotesByLinked, fmt.Sprintf("%s_%d", name, id))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanNote)
}

func (n *Notes) Add(note *models.Note) (models.NoteID, *errors.MasterError) {
	verr := note.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	result, err := n.DB.Exec(SQLAddNote, note.Level, note.Content, note.Linked)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.NoteID(id), nil
}

func (n *Notes) Update(note *models.Note) *errors.MasterError {
	verr := note.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	_, err := n.DB.Exec(SQLUpdateNote, note.Level, note.Content, note.Linked, note.ID)
	return errors.NewMasterError(err, 0)
}

func (n *Notes) Delete(id models.NoteID) *errors.MasterError {
	_, err := n.DB.Exec(SQLDeleteNote, id)
	return errors.NewMasterError(err, 0)
}

func MapNotes(notes []*models.Note) map[models.NoteID]*models.Note {
	noteMap := make(map[models.NoteID]*models.Note)

	for _, note := range notes {
		noteMap[note.ID] = note
	}

	return noteMap
}
