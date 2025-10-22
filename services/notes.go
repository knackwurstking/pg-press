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

	if err := base.CreateTable(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			linked TEXT,
			PRIMARY KEY("id" AUTOINCREMENT)
		);`, TableNameNotes),
		TableNameNotes,
	); err != nil {
		panic(err)
	}

	return &Notes{
		Base: base,
	}
}

func (n *Notes) List() ([]*models.Note, error) {
	n.Log.Debug("Listing notes")

	rows, err := n.DB.Query(
		fmt.Sprintf(
			`SELECT
				id, level, content, created_at, COALESCE(linked, '') as linked
			FROM
				%s;`,
			TableNameNotes,
		),
	)
	if err != nil {
		return nil, n.GetSelectError(err)
	}
	defer rows.Close()

	notes, err := ScanRows(rows, scanNote)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (n *Notes) Get(id int64) (*models.Note, error) {
	n.Log.Debug("Getting note with ID %d", id)

	row := n.DB.QueryRow(
		fmt.Sprintf(`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			%s
		WHERE id = $1;`, TableNameNotes),
		id,
	)

	note, err := ScanSingleRow(row, scanNote)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (n *Notes) GetByIDs(ids []int64) ([]*models.Note, error) {
	n.Log.Debug("Getting notes by IDs (%d): %#v", len(ids), ids)

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
				%[1]s
			WHERE
				id
			IN (%[2]s);`,
			TableNameNotes, strings.Join(placeholders, ","),
		),
		args...,
	)
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
	var notes []*models.Note
	for _, id := range ids {
		if note, exists := noteMap[id]; exists {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

func (n *Notes) GetByPress(press models.PressNumber) ([]*models.Note, error) {
	n.Log.Debug("Getting notes for press: %d", press)

	rows, err := n.DB.Query(
		fmt.Sprintf(`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			%s
		WHERE
			linked = $1;`, TableNameNotes),
		fmt.Sprintf("press_%d", press),
	)
	if err != nil {
		return nil, n.GetSelectError(err)
	}
	defer rows.Close()

	notes, err := ScanRows(rows, scanNote)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (n *MetalSheets) GetByTool(toolID int64) ([]*models.Note, error) {
	n.Log.Debug("Getting notes for tool: %d", toolID)

	rows, err := n.DB.Query(
		fmt.Sprintf(`SELECT
			id, level, content, created_at, COALESCE(linked, '') as linked
		FROM
			%s
		WHERE
			linked = $1;`, TableNameNotes),
		fmt.Sprintf("tool_%d", toolID),
	)
	if err != nil {
		return nil, n.GetSelectError(err)
	}
	defer rows.Close()

	notes, err := ScanRows(rows, scanNote)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (n *Notes) Add(note *models.Note) (int64, error) {
	n.Log.Debug("Adding note: level: %d", note.Level)

	if err := note.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (level, content, linked) VALUES ($1, $2, $3);
	`, TableNameNotes)

	result, err := n.DB.Exec(query, note.Level, note.Content, note.Linked)
	if err != nil {
		return 0, n.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, n.GetInsertError(err)
	}

	return id, nil
}

func (n *Notes) Update(note *models.Note) error {
	n.Log.Debug("Updating note: id: %d, level: %d", note.ID, note.Level)

	if err := note.Validate(); err != nil {
		return err
	}

	_, err := n.DB.Exec(
		fmt.Sprintf(`UPDATE %s SET level = $1, content = $2, linked = $3 WHERE id = $4;`, TableNameNotes),
		note.Level, note.Content, note.Linked, note.ID,
	)
	if err != nil {
		return n.GetUpdateError(err)
	}

	return nil
}

func (n *Notes) Delete(id int64) error {
	n.Log.Debug("Deleting note for ID: %d", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, TableNameNotes)
	_, err := n.DB.Exec(query, id)
	if err != nil {
		return n.GetDeleteError(err)
	}

	return nil
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
