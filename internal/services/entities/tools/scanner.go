package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ScanTool scans a database row into a Tool model
func ScanTool(scannable interfaces.Scannable) (*models.Tool, error) {
	tool := &models.Tool{}
	var format []byte

	err := scannable.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.IsDead, &tool.Press, &tool.Binding)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan tool: %v", err)
	}

	// Unmarshal the format data
	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool format data: %v", err)
	}

	return tool, nil
}

// ScanToolsFromRows scans multiple tool rows
func ScanToolsFromRows(rows *sql.Rows) ([]*models.Tool, error) {
	return scanner.ScanRows(rows, ScanTool)
}

// ScanToolsIntoMap scans tools into a map by ID
func ScanToolsIntoMap(rows *sql.Rows) (map[int64]*models.Tool, error) {
	return scanner.ScanIntoMap(rows, ScanTool, func(tool *models.Tool) int64 {
		return tool.ID
	})
}
