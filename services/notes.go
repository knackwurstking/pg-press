package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
)

const TableNameNotes = "notes"

type Notes struct {
	*Base
}

func NewNotes(r *Registry) *Notes {
	base := NewBase(r, logger.NewComponentLogger("Service: Notes"))

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			linked TEXT,
			PRIMARY KEY("id" AUTOINCREMENT)
		);`, TableNameNotes)

	if err := base.CreateTable(query, TableNameNotes); err != nil {
		panic(err)
	}

	return &Notes{Base: base}
}

func (n *Notes) List() ([]*models.Note, error) {
	n.Log.Debug("Listing notes")

	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s;`,
		TableNameNotes,
	)

	rows, err := n.DB.Query(query)
	if err != nil {
		return nil, n.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanNote)
}

func (n *Notes) Get(id int64) (*models.Note, error) {
	n.Log.Debug("Getting note with ID %d", id)

	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s
		WHERE id = $1;`,
		TableNameNotes,
	)

	row := n.DB.QueryRow(query, id)
	return ScanSingleRow(row, scanNote)
}

func (n *Notes) GetByIDs(ids []int64) ([]*models.Note, error) {
	n.Log.Debug("Getting notes by IDs (%d): %#v", len(ids), ids)

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
		return nil, n.GetSelectError(err)
	}
	defer rows.Close()

	// Store notes in a map for efficient lookup
	noteMap, err := scanNotesIntoMap(rows)
	if err != nil {
		return nil, err
	}

	// Return notes in the order of the requested IDs
	notes := make([]*models.Note, 0, len(ids))
	for _, id := range ids {
		if note, exists := noteMap[id]; exists {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

func (n *Notes) GetByPress(press models.PressNumber) ([]*models.Note, error) {
	n.Log.Debug("Getting notes for press: %d", press)
	return n.getByLinked(fmt.Sprintf("press_%d", press))
}

func (n *Notes) GetByTool(toolID int64) ([]*models.Note, error) {
	n.Log.Debug("Getting notes for tool: %d", toolID)
	return n.getByLinked(fmt.Sprintf("tool_%d", toolID))
}

func (n *Notes) getByLinked(linked string) ([]*models.Note, error) {
	query := fmt.Sprintf(
		`SELECT id, level, content, created_at, COALESCE(linked, '') as linked
		FROM %s
		WHERE linked = $1;`,
		TableNameNotes,
	)

	rows, err := n.DB.Query(query, linked)
	if err != nil {
		return nil, n.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanNote)
}

func (n *Notes) Add(note *models.Note) (int64, error) {
	n.Log.Debug("Adding note: level: %d", note.Level)

	if err := note.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (level, content, linked) VALUES ($1, $2, $3);`,
		TableNameNotes,
	)

	result, err := n.DB.Exec(query, note.Level, note.Content, note.Linked)
	if err != nil {
		return 0, n.GetInsertError(err)
	}

	return result.LastInsertId()
}

func (n *Notes) Update(note *models.Note) error {
	n.Log.Debug("Updating note: id: %d, level: %d", note.ID, note.Level)

	if err := note.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`UPDATE %s SET level = $1, content = $2, linked = $3 WHERE id = $4;`,
		TableNameNotes,
	)

	_, err := n.DB.Exec(query, note.Level, note.Content, note.Linked, note.ID)
	return n.GetUpdateError(err)
}

func (n *Notes) Delete(id int64) error {
	n.Log.Debug("Deleting note for ID: %d", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, TableNameNotes)
	_, err := n.DB.Exec(query, id)
	return n.GetDeleteError(err)
}

func scanNote(scanner Scannable) (*models.Note, error) {
	note := &models.Note{}
	err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt, &note.Linked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan note: %v", err)
	}
	return note, nil
}

func scanNotesIntoMap(rows *sql.Rows) (map[int64]*models.Note, error) {
	resultMap := make(map[int64]*models.Note)

	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		resultMap[note.ID] = note
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return resultMap, nil
}
