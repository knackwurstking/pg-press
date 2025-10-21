package presscycles

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ScanPressCycle scans a database row into a Cycle model
func ScanPressCycle(scannable interfaces.Scannable) (*models.Cycle, error) {
	cycle := &models.Cycle{}
	var performedBy sql.NullInt64

	err := scannable.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.ToolPosition,
		&cycle.TotalCycles,
		&cycle.Date,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan press cycle: %v", err)
	}

	if performedBy.Valid {
		cycle.PerformedBy = performedBy.Int64
	}

	return cycle, nil
}

// ScanPressCyclesFromRows scans multiple press cycle rows
func ScanPressCyclesFromRows(rows *sql.Rows) ([]*models.Cycle, error) {
	return scanner.ScanRows(rows, ScanPressCycle)
}

// ScanPressCyclesIntoMap scans press cycles into a map by ID
func ScanPressCyclesIntoMap(rows *sql.Rows) (map[int64]*models.Cycle, error) {
	return scanner.ScanIntoMap(rows, ScanPressCycle, func(cycle *models.Cycle) int64 {
		return cycle.ID
	})
}
