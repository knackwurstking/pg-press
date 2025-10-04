package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Notes struct {
	db  *sql.DB
	log *logger.Logger
}

func NewNotes(db *sql.DB) *Notes {
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

	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create notes table: %v", err))
	}

	return &Notes{
		db:  db,
		log: logger.GetComponentLogger("Service: Notes"),
	}
}

func (n *Notes) List() ([]*models.Note, error) {
	n.log.Info("Listing notes")

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes;
	`

	rows, err := n.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: notes: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %v", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: notes: %v", err)
	}

	return notes, nil
}

func (n *Notes) Get(id int64) (*models.Note, error) {
	n.log.Info("Getting note, id: %d", id)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked FROM notes WHERE id = $1;
	`

	row := n.db.QueryRow(query, id)

	note, err := n.scanNote(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("note with ID %d not found", id))
		}
		return nil, fmt.Errorf("select error: notes: %v", err)
	}

	return note, nil
}

func (n *Notes) GetByIDs(ids []int64) ([]*models.Note, error) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}

	n.log.Debug("Getting notes by IDs: %v", ids)

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
		return nil, fmt.Errorf("select error: notes: %v", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	noteMap := make(map[int64]*models.Note)

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %v", err)
		}
		noteMap[note.ID] = note
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: notes: %v", err)
	}

	// Return attachments in the order of the requested IDs
	var notes []*models.Note
	for _, id := range ids {
		if note, exists := noteMap[id]; exists {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

func (n *Notes) GetByPress(press models.PressNumber) ([]*models.Note, error) {
	n.log.Debug("Getting notes for press %d", press)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM notes
		WHERE linked = $1;
	`

	pressLinked := fmt.Sprintf("press_%d", press)
	rows, err := n.db.Query(query, pressLinked)
	if err != nil {
		return nil, fmt.Errorf("select error: notes by press: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %v", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: notes by press: %v", err)
	}

	return notes, nil
}

func (n *Notes) GetByTool(toolID int64) ([]*models.Note, error) {
	n.log.Debug("Getting notes for tool %d", toolID)

	query := `
		SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM notes
		WHERE linked = $1;
	`

	toolLinked := fmt.Sprintf("tool_%d", toolID)
	rows, err := n.db.Query(query, toolLinked)
	if err != nil {
		return nil, fmt.Errorf("select error: notes by tool: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %v", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: notes by tool: %v", err)
	}

	return notes, nil
}

func (n *Notes) Add(note *models.Note) (int64, error) {
	n.log.Info("Adding note: level=%d", note.Level)

	query := `
		INSERT INTO notes (level, content, linked) VALUES ($1, $2, $3);
	`

	// Convert empty string to NULL for database storage
	var linkedValue interface{}
	if note.Linked == "" {
		linkedValue = nil
	} else {
		linkedValue = note.Linked
	}

	result, err := n.db.Exec(query, note.Level, note.Content, linkedValue)
	if err != nil {
		return 0, fmt.Errorf("insert error: notes: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert error: notes: %v", err)
	}

	return id, nil
}

func (n *Notes) Update(note *models.Note) error {
	n.log.Info("Updating note: id=%d, level=%d", note.ID, note.Level)

	query := `
		UPDATE notes SET level = $1, content = $2, linked = $3 WHERE id = $4;
	`

	// Convert empty string to NULL for database storage
	var linkedValue interface{}
	if note.Linked == "" {
		linkedValue = nil
	} else {
		linkedValue = note.Linked
	}

	result, err := n.db.Exec(query, note.Level, note.Content, linkedValue, note.ID)
	if err != nil {
		return fmt.Errorf("update error: notes: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with id %d not found", note.ID)
	}

	return nil
}

func (n *Notes) Delete(id int64, user *models.User) error {
	n.log.Info("Deleting note: id=%d by user=%s", id, user.Name)

	query := `
		DELETE FROM notes WHERE id = $1;
	`

	result, err := n.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete error: notes: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError(fmt.Sprintf("note with ID %d not found", id))
	}

	return nil
}

func (n *Notes) scanNote(scanner interfaces.Scannable) (*models.Note, error) {
	note := &models.Note{}

	if err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt, &note.Linked); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}

		return nil, fmt.Errorf("scan error: notes: %v", err)
	}

	return note, nil
}
