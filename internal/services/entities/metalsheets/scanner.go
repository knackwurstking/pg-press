package metalsheets

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func scanMetalSheet(scanner interfaces.Scannable) (*models.MetalSheet, error) {
	sheet := &models.MetalSheet{}
	var identifierStr string
	var toolID int64

	err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight,
		&sheet.STF, &sheet.STFMax, &identifierStr, &toolID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan metal sheet: %v", err)
	}

	// Convert string identifier to MachineType
	sheet.Identifier = models.MachineType(identifierStr)
	sheet.ToolID = toolID

	return sheet, nil
}

func scanMetalSheetsFromRows(rows *sql.Rows) ([]*models.MetalSheet, error) {
	return scanner.ScanRows(rows, scanMetalSheet)
}
