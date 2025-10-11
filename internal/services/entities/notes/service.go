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

	return &Service{
		BaseService: base,
	}
}

func (n *Service) List() ([]*models.Note, error) {
	n.Log.Debug("Listing notes")

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes;
	`

	rows, err := n.DB.Query(query)
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
	if err := validation.ValidateID(id, "note"); err != nil {
		return nil, err
	}

	n.Log.Debug("Getting note: %d", id)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes WHERE id = $1;
	`

	row := n.DB.QueryRow(query, id)

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

	n.Log.Debug("Getting notes by IDs: count: %d", len(ids))

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

	rows, err := n.DB.Query(query, args...)
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

	n.Log.Debug("Found notes by IDs: count: %d", len(notes))
	return notes, nil
}

func (n *Service) GetByPress(press models.PressNumber) ([]*models.Note, error) {
	n.Log.Debug("Getting notes for press: %d", press)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM notes
		WHERE linked = $1;
	`

	pressLinked := fmt.Sprintf("press_%d", press)
	rows, err := n.DB.Query(query, pressLinked)
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
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	n.Log.Debug("Getting notes for tool: %d", toolID)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM notes
		WHERE linked = $1;
	`

	toolLinked := fmt.Sprintf("tool_%d", toolID)
	rows, err := n.DB.Query(query, toolLinked)
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
	if err := validateNote(note); err != nil {
		return 0, err
	}

	n.Log.Debug("Adding note: level: %d", note.Level)

	query := `
		INSERT INTO notes (level, content, linked) VALUES ($1, $2, $3);
	`

	// Convert empty string to NULL for database storage
	linkedValue := n.PrepareNullableString(note.Linked)

	result, err := n.DB.Exec(query, note.Level, note.Content, linkedValue)
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
	if err := validateNote(note); err != nil {
		return err
	}

	if err := validation.ValidateID(note.ID, "note"); err != nil {
		return err
	}

	n.Log.Debug("Updating note: id: %d, level: %d", note.ID, note.Level)

	query := `
		UPDATE notes SET level = $1, content = $2, linked = $3 WHERE id = $4;
	`

	// Convert empty string to NULL for database storage
	linkedValue := n.PrepareNullableString(note.Linked)

	result, err := n.DB.Exec(query, note.Level, note.Content, linkedValue, note.ID)
	if err != nil {
		return n.HandleUpdateError(err, "notes")
	}

	if err := n.CheckRowsAffected(result, "note", note.ID); err != nil {
		return err
	}

	n.Log.Debug("Updated note: id: %d", note.ID)
	return nil
}

func (n *Service) Delete(id int64, user *models.User) error {
	if err := validation.ValidateID(id, "note"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	n.Log.Debug("Deleting note: id: %d: user: %s", id, user)

	query := `
		DELETE FROM notes WHERE id = $1;
	`

	result, err := n.DB.Exec(query, id)
	if err != nil {
		return n.HandleDeleteError(err, "notes")
	}

	if err := n.CheckRowsAffected(result, "note", id); err != nil {
		return err
	}

	return nil
}
