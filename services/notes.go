package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameNotes = "notes"

type Notes struct {
	*Base
}

func NewNotes(r *Registry) *Notes {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			linked TEXT,
			PRIMARY KEY("id" AUTOINCREMENT)
		);`, TableNameNotes)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameNotes))
	}

	return &Notes{Base: base}
}

func (n *Notes) List() ([]*models.Note, *errors.MasterError) {
	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s;`,
		TableNameNotes,
	)

	rows, err := n.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanNote)
}

func (n *Notes) Get(id models.NoteID) (*models.Note, *errors.MasterError) {
	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s
		WHERE id = $1;`,
		TableNameNotes,
	)

	row := n.DB.QueryRow(query, id)
	note, err := ScanNote(row)
	if err != nil {
		return note, errors.NewMasterError(err, 0)
	}
	return note, nil
}

func (n *Notes) GetByIDs(ids []models.NoteID) ([]*models.Note, *errors.MasterError) {
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

	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s
		WHERE id IN (%s);`,
		TableNameNotes,
		strings.Join(placeholders, ","),
	)

	rows, err := n.DB.Query(query, args...)
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

func (n *Notes) GetByPress(press models.PressNumber) ([]*models.Note, *errors.MasterError) {
	return n.getByLinked(fmt.Sprintf("press_%d", press))
}

func (n *Notes) GetByTool(toolID models.ToolID) ([]*models.Note, *errors.MasterError) {
	return n.getByLinked(fmt.Sprintf("tool_%d", toolID))
}

func (n *Notes) getByLinked(linked string) ([]*models.Note, *errors.MasterError) {
	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s
		WHERE linked = $1;`,
		TableNameNotes,
	)

	rows, err := n.DB.Query(query, linked)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanNote)
}

func (n *Notes) Add(note *models.Note) (models.NoteID, *errors.MasterError) {
	if !note.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid note: %v", note), http.StatusBadRequest)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (level, content, linked) VALUES ($1, $2, $3);`,
		TableNameNotes,
	)

	result, err := n.DB.Exec(query, note.Level, note.Content, note.Linked)
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
	if !note.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid note: %v", note), http.StatusBadRequest)
	}

	query := fmt.Sprintf(
		`UPDATE %s SET level = $1, content = $2, linked = $3 WHERE id = $4;`,
		TableNameNotes,
	)

	_, err := n.DB.Exec(query, note.Level, note.Content, note.Linked, note.ID)
	return errors.NewMasterError(err, 0)
}

func (n *Notes) Delete(id models.NoteID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, TableNameNotes)
	_, err := n.DB.Exec(query, id)
	return errors.NewMasterError(err, 0)
}

func MapNotes(notes []*models.Note) map[models.NoteID]*models.Note {
	noteMap := make(map[models.NoteID]*models.Note)

	for _, note := range notes {
		noteMap[note.ID] = note
	}

	return noteMap
}
