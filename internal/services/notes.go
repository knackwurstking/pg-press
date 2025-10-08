package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Notes struct {
	*BaseService
}

func NewNotes(db *sql.DB) *Notes {
	base := NewBaseService(db, "Notes")

	query := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			linked TEXT,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if err := base.CreateTable(query, "notes"); err != nil {
		panic(err)
	}

	return &Notes{
		BaseService: base,
	}
}

func (n *Notes) List() ([]*models.Note, error) {
	n.LogOperation("Listing notes")

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes;
	`

	rows, err := n.db.Query(query)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes")
	}
	defer rows.Close()

	notes, err := ScanNotesFromRows(rows)
	if err != nil {
		return nil, err
	}

	n.LogOperation("Listed notes", fmt.Sprintf("count: %d", len(notes)))
	return notes, nil
}

func (n *Notes) Get(id int64) (*models.Note, error) {
	if err := ValidateID(id, "note"); err != nil {
		return nil, err
	}

	n.LogOperation("Getting note", id)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes WHERE id = $1;
	`

	row := n.db.QueryRow(query, id)

	note, err := ScanSingleRow(row, ScanNote, "notes")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("note with ID %d not found", id))
		}
		return nil, err
	}

	return note, nil
}

func (n *Notes) GetByIDs(ids []int64) ([]*models.Note, error) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}

	n.LogOperation("Getting notes by IDs", fmt.Sprintf("count: %d", len(ids)))

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes WHERE id IN (%s);`,
		strings.Join(placeholders, ","),
	)

	rows, err := n.db.Query(query, args...)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes")
	}
	defer rows.Close()

	// Store notes in a map for efficient lookup
	noteMap, err := ScanNotesIntoMap(rows)
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

	n.LogOperation("Found notes by IDs", fmt.Sprintf("found: %d", len(notes)))
	return notes, nil
}

func (n *Notes) GetByPress(press models.PressNumber) ([]*models.Note, error) {
	n.LogOperation("Getting notes for press", press)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM notes
		WHERE linked = $1;
	`

	pressLinked := fmt.Sprintf("press_%d", press)
	rows, err := n.db.Query(query, pressLinked)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes by press")
	}
	defer rows.Close()

	notes, err := ScanNotesFromRows(rows)
	if err != nil {
		return nil, err
	}

	n.LogOperation("Found notes for press", fmt.Sprintf("press: %d, count: %d", press, len(notes)))
	return notes, nil
}

func (n *Notes) GetByTool(toolID int64) ([]*models.Note, error) {
	if err := ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	n.LogOperation("Getting notes for tool", toolID)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM notes
		WHERE linked = $1;
	`

	toolLinked := fmt.Sprintf("tool_%d", toolID)
	rows, err := n.db.Query(query, toolLinked)
	if err != nil {
		return nil, n.HandleSelectError(err, "notes by tool")
	}
	defer rows.Close()

	notes, err := ScanNotesFromRows(rows)
	if err != nil {
		return nil, err
	}

	n.LogOperation("Found notes for tool", fmt.Sprintf("tool: %d, count: %d", toolID, len(notes)))
	return notes, nil
}

func (n *Notes) Add(note *models.Note) (int64, error) {
	if err := ValidateNote(note); err != nil {
		return 0, err
	}

	n.LogOperation("Adding note", fmt.Sprintf("level: %d", note.Level))

	query := `
		INSERT INTO notes (level, content, linked) VALUES ($1, $2, $3);
	`

	// Convert empty string to NULL for database storage
	linkedValue := n.PrepareNullableString(note.Linked)

	result, err := n.db.Exec(query, note.Level, note.Content, linkedValue)
	if err != nil {
		return 0, n.HandleInsertError(err, "notes")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, n.HandleInsertError(err, "notes")
	}

	n.LogOperation("Added note", fmt.Sprintf("id: %d", id))
	return id, nil
}

func (n *Notes) Update(note *models.Note) error {
	if err := ValidateNote(note); err != nil {
		return err
	}

	if err := ValidateID(note.ID, "note"); err != nil {
		return err
	}

	n.LogOperation("Updating note", fmt.Sprintf("id: %d, level: %d", note.ID, note.Level))

	query := `
		UPDATE notes SET level = $1, content = $2, linked = $3 WHERE id = $4;
	`

	// Convert empty string to NULL for database storage
	linkedValue := n.PrepareNullableString(note.Linked)

	result, err := n.db.Exec(query, note.Level, note.Content, linkedValue, note.ID)
	if err != nil {
		return n.HandleUpdateError(err, "notes")
	}

	if err := n.CheckRowsAffected(result, "note", note.ID); err != nil {
		return err
	}

	n.LogOperation("Updated note", fmt.Sprintf("id: %d", note.ID))
	return nil
}

func (n *Notes) Delete(id int64, user *models.User) error {
	if err := ValidateID(id, "note"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	n.LogOperationWithUser("Deleting note", createUserInfo(user), fmt.Sprintf("id: %d", id))

	query := `
		DELETE FROM notes WHERE id = $1;
	`

	result, err := n.db.Exec(query, id)
	if err != nil {
		return n.HandleDeleteError(err, "notes")
	}

	if err := n.CheckRowsAffected(result, "note", id); err != nil {
		return err
	}

	return nil
}
