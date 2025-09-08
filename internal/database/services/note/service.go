package note

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	notemodels "github.com/knackwurstking/pgpress/internal/database/models/note"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	dbutils "github.com/knackwurstking/pgpress/internal/database/utils"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db *sql.DB
}

var _ interfaces.DataOperations[*notemodels.Note] = (*Service)(nil)

func New(db *sql.DB) *Service {
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
		panic(
			dberror.NewDatabaseError(
				"create_table",
				"notes",
				"failed to create notes table",
				err,
			),
		)
	}

	return &Service{
		db: db,
	}
}

func (n *Service) List() ([]*notemodels.Note, error) {
	logger.DBNotes().Info("Listing notes")

	query := `
		SELECT id, level, content, created_at FROM notes;
	`

	rows, err := n.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "notes",
			"failed to query notes", err)
	}
	defer rows.Close()

	var notes []*notemodels.Note

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan note")
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "notes",
			"error iterating over rows", err)
	}

	return notes, nil
}

func (n *Service) Get(id int64) (*notemodels.Note, error) {
	logger.DBNotes().Info("Getting note, id: %d", id)

	query := `
		SELECT id, level, content, created_at FROM notes WHERE id = $1;
	`

	row := n.db.QueryRow(query, id)

	note, err := n.scanNote(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "notes",
			fmt.Sprintf("failed to get note with ID %d", id), err)
	}

	return note, nil
}

func (n *Service) GetByIDs(ids []int64) ([]*notemodels.Note, error) {
	if len(ids) == 0 {
		return []*notemodels.Note{}, nil
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
		dbutils.JoinStrings(placeholders, ","),
	)

	rows, err := n.db.Query(query, args...)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "notes",
			"failed to query notes by IDs", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	noteMap := make(map[int64]*notemodels.Note)

	for rows.Next() {
		note, err := n.scanNote(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan attachment")
		}
		noteMap[note.ID] = note
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "notes",
			"error iterating over rows", err)
	}

	// Return attachments in the order of the requested IDs
	var notes []*notemodels.Note
	for _, id := range ids {
		if note, exists := noteMap[id]; exists {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

func (n *Service) Add(note *notemodels.Note, _ *usermodels.User) (int64, error) {
	logger.DBNotes().Info("Adding note: level=%d", note.Level)

	query := `
		INSERT INTO notes (level, content) VALUES ($1, $2);
	`

	result, err := n.db.Exec(query, note.Level, note.Content)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "notes",
			"failed to insert note", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "notes",
			"failed to get last insert ID", err)
	}

	return id, nil
}

func (n *Service) Update(note *notemodels.Note, user *usermodels.User) error {
	return fmt.Errorf("operation not supported")
}

func (n *Service) Delete(id int64, user *usermodels.User) error {
	return fmt.Errorf("operation not supported")
}

func (n *Service) scanNote(scanner interfaces.Scannable) (*notemodels.Note, error) {
	note := &notemodels.Note{}

	if err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "notes",
			"failed to scan row", err)
	}

	return note, nil
}
