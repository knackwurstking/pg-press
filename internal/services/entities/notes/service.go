package notes

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
}

func NewService(db *sql.DB) *Service {
	base := base.NewBaseService(db, "Notes")

	if err := base.CreateTable(
		`CREATE TABLE IF NOT EXISTS notes (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			linked TEXT,
			PRIMARY KEY("id" AUTOINCREMENT)
		);`,
		"notes",
	); err != nil {
		panic(err)
	}

	return &Service{
		BaseService: base,
	}
}

func (n *Service) List() ([]*models.Note, error) {
	rows, err := n.DB.Query(
		`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			notes;`,
	)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes")
	}
	defer rows.Close()

	notes, err := scanNotesFromRows(rows)
	if err != nil {
		return nil, err
	}

	n.Log.Debug("Listed notes: count: %d", len(notes))
	return notes, nil
}

func (n *Service) Get(id int64) (*models.Note, error) {
	row := n.DB.QueryRow(
		`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			notes
		WHERE id = $1;`,
		id,
	)

	note, err := scanner.ScanSingleRow(row, scanNote, "notes")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("note with ID %d not found", id))
		}
		return nil, err
	}

	return note, nil
}

func (n *Service) GetByIDs(ids []int64) ([]*models.Note, error) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids)) // Yeah, need to convert this []int64 to []any
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	rows, err := n.DB.Query(
		fmt.Sprintf(
			`SELECT
				id, level, content, created_at, COALESCE(linked, '') as linked
			FROM
				notes
			WHERE
				id
			IN (%s);`,
			strings.Join(placeholders, ","),
		),
		args...,
	)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes")
	}
	defer rows.Close()

	// Store notes in a map for efficient lookup
	noteMap, err := scanNotesIntoMap(rows)
	if err != nil {
		return nil, err
	}

	// Return notes in the order of the requested IDs
	var notes []*models.Note
	for _, id := range ids {
		if note, exists := noteMap[id]; exists {
			notes = append(notes, note)
		}
	}

	n.Log.Debug("Found notes by IDs (%d): count: %d", len(ids), len(notes))
	return notes, nil
}

func (n *Service) GetByPress(press models.PressNumber) ([]*models.Note, error) {
	rows, err := n.DB.Query(
		`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			notes
		WHERE
			linked = $1;`,
		fmt.Sprintf("press_%d", press),
	)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes by press")
	}
	defer rows.Close()

	notes, err := scanNotesFromRows(rows)
	if err != nil {
		return nil, err
	}

	n.Log.Debug("Found notes for press: press: %d, count: %d", press, len(notes))
	return notes, nil
}

func (n *Service) GetByTool(toolID int64) ([]*models.Note, error) {
	rows, err := n.DB.Query(
		`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			notes
		WHERE
			linked = $1;`,
		fmt.Sprintf("tool_%d", toolID),
	)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes by tool")
	}
	defer rows.Close()

	notes, err := scanNotesFromRows(rows)
	if err != nil {
		return nil, err
	}

	n.Log.Debug("Found notes for tool: tool: %d, count: %d", toolID, len(notes))
	return notes, nil
}

func (n *Service) Add(note *models.Note) (int64, error) {
	err := validation.ValidateNotNil(note, "note")
	if err != nil {
		return 0, err
	}

	err = note.Validate()
	if err != nil {
		return 0, err
	}

	n.Log.Debug("Adding note: level: %d", note.Level)

	query := `
		INSERT INTO notes (level, content, linked) VALUES ($1, $2, $3);
	`

	result, err := n.DB.Exec(query, note.Level, note.Content, note.Linked)
	if err != nil {
		return 0, n.HandleInsertError(err, "notes")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, n.HandleInsertError(err, "notes")
	}

	n.Log.Debug("Added note: id: %d", id)
	return id, nil
}

func (n *Service) Update(note *models.Note) error {
	err := validation.ValidateNotNil(note, "note")
	if err != nil {
		return err
	}

	err = note.Validate()
	if err != nil {
		return err
	}

	n.Log.Debug("Updating note: id: %d, level: %d", note.ID, note.Level)

	result, err := n.DB.Exec(
		`UPDATE notes SET level = $1, content = $2, linked = $3 WHERE id = $4;`,
		note.Level, note.Content, note.Linked, note.ID,
	)
	if err != nil {
		return n.HandleUpdateError(err, "notes")
	}

	if err := n.CheckRowsAffected(result, "note", note.ID); err != nil {
		return err
	}

	return nil
}

func (n *Service) Delete(id int64) error {
	n.Log.Debug("Deleting note for ID: %d", id)

	result, err := n.DB.Exec(`DELETE FROM notes WHERE id = $1;`, id)
	if err != nil {
		return n.HandleDeleteError(err, "notes")
	}

	if err := n.CheckRowsAffected(result, "note", id); err != nil {
		return err
	}

	return nil
}
