package toolregenerations

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ScanToolRegeneration scans a database row into a Regeneration model
func ScanToolRegeneration(scannable interfaces.Scannable) (*models.Regeneration, error) {
	regen := &models.Regeneration{}
	var performedBy sql.NullInt64

	err := scannable.Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.CycleID,
		&regen.Reason,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan tool regeneration: %v", err)
	}

	if performedBy.Valid {
		regen.PerformedBy = &performedBy.Int64
	}

	return regen, nil
}

// ScanToolRegenerationsFromRows scans multiple tool regeneration rows
func ScanToolRegenerationsFromRows(rows *sql.Rows) ([]*models.Regeneration, error) {
	return scanner.ScanRows(rows, ScanToolRegeneration)
}

// ScanToolRegenerationsIntoMap scans tool regenerations into a map by ID
func ScanToolRegenerationsIntoMap(rows *sql.Rows) (map[int64]*models.Regeneration, error) {
	return scanner.ScanIntoMap(rows, ScanToolRegeneration, func(regen *models.Regeneration) int64 {
		return regen.ID
	})
}
