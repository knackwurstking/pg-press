// Service for managing metal sheets with validation
package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
)

const TableNameMetalSheets = "metal_sheets"

type MetalSheets struct {
	*Base
}

func NewMetalSheets(r *Registry) *MetalSheets {
	base := NewBase(r, logger.NewComponentLogger("Service: MetalSheets"))

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			tile_height REAL NOT NULL,
			value REAL NOT NULL,
			marke_height INTEGER NOT NULL,
			stf REAL NOT NULL,
			stf_max REAL NOT NULL,
			identifier TEXT NOT NULL,
			tool_id INTEGER NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY("id" AUTOINCREMENT),
			FOREIGN KEY("tool_id") REFERENCES "tools"("id") ON DELETE CASCADE
		);
	`, TableNameMetalSheets)

	if err := base.CreateTable(query, TableNameMetalSheets); err != nil {
		panic(err)
	}

	return &MetalSheets{
		Base: base,
	}
}

func (s *MetalSheets) List() ([]*models.MetalSheet, error) {
	s.Log.Debug("Listing metal sheets")

	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		ORDER BY id DESC;
	`, TableNameMetalSheets)

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	sheets, err := ScanRows(rows, scanMetalSheet)
	if err != nil {
		return nil, err
	}

	return sheets, nil
}

func (s *MetalSheets) Get(id int64) (*models.MetalSheet, error) {
	s.Log.Debug("Getting metal sheet: %d", id)

	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		WHERE id = ?;
	`, TableNameMetalSheets)

	row := s.DB.QueryRow(query, id)

	sheet, err := ScanSingleRow(row, scanMetalSheet)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(
				fmt.Sprintf("metal sheet with ID %d", id))
		}

		return nil, err
	}

	return sheet, nil
}

func (s *MetalSheets) GetByToolID(toolID int64) ([]*models.MetalSheet, error) {
	s.Log.Debug("Getting metal sheets for tool: %d", toolID)

	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		WHERE tool_id = $1
		ORDER BY id DESC;
	`, TableNameMetalSheets)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	sheets, err := ScanRows(rows, scanMetalSheet)
	if err != nil {
		return nil, err
	}

	return sheets, nil
}

func (s *MetalSheets) GetByMachineType(machineType models.MachineType) ([]*models.MetalSheet, error) {
	s.Log.Debug("Getting metal sheets for machine type: %s", string(machineType))

	if !machineType.IsValid() {
		return nil, errors.NewValidationError(
			fmt.Sprintf("invalid machine type: %s", machineType))
	}

	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		WHERE identifier = $1
		ORDER BY id DESC;
	`, TableNameMetalSheets)

	rows, err := s.DB.Query(query, machineType.String())
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	sheets, err := ScanRows(rows, scanMetalSheet)
	if err != nil {
		return nil, err
	}

	return sheets, nil
}

func (s *MetalSheets) GetForPress(pressNumber models.PressNumber, toolsMap map[int64]*models.Tool) ([]*models.MetalSheet, error) {
	s.Log.Debug("Getting metal sheets for press: %d, tools: %d", pressNumber, len(toolsMap))

	// Get the expected machine type for this press
	expectedMachineType := models.GetMachineTypeForPress(pressNumber)
	s.Log.Debug("Press machine type determined: press: %d, type: %s",
		pressNumber, expectedMachineType)

	// Get metal sheets for all tools on this press
	var allSheets models.MetalSheetList
	for toolID := range toolsMap {
		sheets, err := s.GetByToolID(toolID)
		if err != nil {
			s.Log.Error("Failed to get metal sheets for tool %d: %v", toolID, err)
			continue
		}
		s.Log.Debug("Retrieved sheets for tool: tool: %d, count: %d", toolID, len(sheets))
		allSheets = append(allSheets, sheets...)
	}

	// Filter by machine type and count by type
	var filteredSheets models.MetalSheetList
	var sacmiCount, sitiCount, otherCount int

	for _, sheet := range allSheets {
		switch sheet.Identifier {
		case models.MachineTypeSACMI:
			sacmiCount++
			if sheet.Identifier == expectedMachineType {
				filteredSheets = append(filteredSheets, sheet)
			}
		case models.MachineTypeSITI:
			sitiCount++
			if sheet.Identifier == expectedMachineType {
				filteredSheets = append(filteredSheets, sheet)
			}
		default:
			otherCount++
			s.Log.Warn("Found metal sheet %d with unexpected identifier: %s", sheet.ID, sheet.Identifier)
		}
	}

	return filteredSheets, nil
}

func (s *MetalSheets) Add(sheet *models.MetalSheet) (int64, error) {
	s.Log.Debug("Adding metal sheet: tool_id: %d, identifier: %s",
		sheet.ToolID, sheet.Identifier)

	// Validate the metal sheet before adding
	if err := sheet.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (tile_height, value, marke_height, stf, stf_max, identifier, tool_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`, TableNameMetalSheets)

	result, err := s.DB.Exec(
		query,
		sheet.TileHeight,
		sheet.Value,
		sheet.MarkeHeight,
		sheet.STF,
		sheet.STFMax,
		sheet.Identifier.String(),
		sheet.ToolID,
	)
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	// Set sheet ID for return
	sheet.ID = id

	return id, nil
}

func (s *MetalSheets) Update(sheet *models.MetalSheet) error {
	s.Log.Debug("Updating metal sheet: %d", sheet.ID)

	// Validate the metal sheet before updating
	if err := sheet.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET tile_height = $1, value = $2, marke_height = $3, stf = $4, stf_max = $5,
			identifier = $6, tool_id = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8;
	`, TableNameMetalSheets)

	_, err := s.DB.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.Identifier.String(), sheet.ToolID, sheet.ID)
	if err != nil {
		return s.GetUpdateError(err)
	}

	return nil
}

func (s *MetalSheets) AssignTool(sheetID int64, toolID int64) error {
	s.Log.Debug("Assigning tool to metal sheet: sheet_id: %d, tool_id: %d", sheetID, toolID)

	// Validate tool ID
	if toolID <= 0 {
		return errors.NewValidationError("tool_id: must be positive")
	}

	// Get current sheet to verify it exists
	_, err := s.Get(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get sheet for tool assignment: %v", err)
	}

	// Update the tool assignment
	_, err = s.DB.Exec(
		fmt.Sprintf(`UPDATE %s
		SET tool_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`, TableNameMetalSheets),
		toolID, sheetID,
	)
	if err != nil {
		return s.GetUpdateError(err)
	}

	return nil
}

func (s *MetalSheets) Delete(id int64) error {
	s.Log.Debug("Deleting metal sheet: %d", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, TableNameMetalSheets)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func scanMetalSheet(scanner Scannable) (*models.MetalSheet, error) {
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
