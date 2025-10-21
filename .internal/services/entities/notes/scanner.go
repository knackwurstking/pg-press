package notes

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func scanNote(scanner interfaces.Scannable) (*models.Note, error) {
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

func scanNotesFromRows(rows *sql.Rows) ([]*models.Note, error) {
	return scanner.ScanRows(rows, scanNote)
}

func scanNotesIntoMap(rows *sql.Rows) (map[int64]*models.Note, error) {
	return scanner.ScanIntoMap(rows, scanNote, func(note *models.Note) int64 {
		return note.ID
	})
}
