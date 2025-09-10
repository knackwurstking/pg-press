package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Note struct {
	db *sql.DB
}

func NewNote(db *sql.DB) *Note {
	query := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create notes table: %w", err))
	}

	return &Note{
		db: db,
	}
}

func (n *Note) List() ([]*models.Note, error) {
	logger.DBNotes().Info("Listing notes")

	query := `
		SELECT id, level, content, created_at FROM notes;
	`

	rows, err := n.db.Query(query)
	if err != nil {
		return nil, utils.NewDatabaseError("select", "notes", err)
	}
	defer rows.Close()

	var notes []*models.Note

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.NewDatabaseError("select", "notes", err)
	}

	return notes, nil
}

func (n *Note) Get(id int64) (*models.Note, error) {
	logger.DBNotes().Info("Getting note, id: %d", id)

	query := `
		SELECT id, level, content, created_at FROM notes WHERE id = $1;
	`

	row := n.db.QueryRow(query, id)

	note, err := n.scanNote(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("note with ID %d not found", id))
		}
		return nil, utils.NewDatabaseError("select", "notes", err)
	}

	return note, nil
}

func (n *Note) GetByIDs(ids []int64) ([]*models.Note, error) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}

	logger.DBNotes().Debug("Getting notes by IDs: %v", ids)

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT id, level, content, created_at FROM notes WHERE id IN (%s);`,
		strings.Join(placeholders, ","),
	)

	rows, err := n.db.Query(query, args...)
	if err != nil {
		return nil, utils.NewDatabaseError("select", "notes", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	noteMap := make(map[int64]*models.Note)

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		noteMap[note.ID] = note
	}

	if err := rows.Err(); err != nil {
		return nil, utils.NewDatabaseError("select", "notes", err)
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

func (n *Note) Add(note *models.Note, _ *models.User) (int64, error) {
	logger.DBNotes().Info("Adding note: level=%d", note.Level)

	query := `
		INSERT INTO notes (level, content) VALUES ($1, $2);
	`

	result, err := n.db.Exec(query, note.Level, note.Content)
	if err != nil {
		return 0, utils.NewDatabaseError("insert", "notes", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, utils.NewDatabaseError("insert", "notes", err)
	}

	return id, nil
}

func (n *Note) Update(note *models.Note, user *models.User) error {
	return fmt.Errorf("operation not supported")
}

func (n *Note) Delete(id int64, user *models.User) error {
	return fmt.Errorf("operation not supported")
}

func (n *Note) scanNote(scanner interfaces.Scannable) (*models.Note, error) {
	note := &models.Note{}

	if err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}

		return nil, utils.NewDatabaseError("scan", "notes", err)
	}

	return note, nil
}
