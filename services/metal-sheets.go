// Service for managing metal sheets with validation
package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameMetalSheets = "metal_sheets"

type MetalSheets struct {
	*Base
}

func NewMetalSheets(r *Registry) *MetalSheets {
	base := NewBase(r)

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

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameMetalSheets))
	}

	return &MetalSheets{
		Base: base,
	}
}

func (s *MetalSheets) List() ([]*models.MetalSheet, error) {
	slog.Debug("Listing metal sheets")

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

	return ScanRows(rows, scanMetalSheet)
}

func (s *MetalSheets) Get(id models.MetalSheetID) (*models.MetalSheet, error) {
	slog.Debug("Getting metal sheet", "id", id)

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

func (s *MetalSheets) GetByToolID(toolID models.ToolID) ([]*models.MetalSheet, error) {
	slog.Debug("Getting metal sheets for tool", "tool", toolID)

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

	return ScanRows(rows, scanMetalSheet)
}

func (s *MetalSheets) GetByMachineType(machineType models.MachineType) ([]*models.MetalSheet, error) {
	slog.Debug("Getting metal sheets for machine type", "machine_type", machineType)

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

	return ScanRows(rows, scanMetalSheet)
}

func (s *MetalSheets) GetForPress(pressNumber models.PressNumber, toolsMap map[models.ToolID]*models.Tool) ([]*models.MetalSheet, error) {
	slog.Debug("Getting metal sheets for press", "press", pressNumber, "tools", len(toolsMap))

	expectedMachineType := models.GetMachineTypeForPress(pressNumber)
	slog.Debug("Press machine type determined",
		"press", pressNumber, "machine-type", expectedMachineType)

	var allSheets models.MetalSheetList
	for toolID := range toolsMap {
		sheets, err := s.GetByToolID(toolID)
		if err != nil {
			slog.Error("Failed to get metal sheets for tool", "tool", toolID, "error", err)
			continue
		}
		slog.Debug("Retrieved sheets for tool", "tool", toolID, "sheets", len(sheets))
		allSheets = append(allSheets, sheets...)
	}

	var filteredSheets models.MetalSheetList
	for _, sheet := range allSheets {
		if sheet.Identifier == expectedMachineType {
			filteredSheets = append(filteredSheets, sheet)
		} else if sheet.Identifier != models.MachineTypeSACMI && sheet.Identifier != models.MachineTypeSITI {
			slog.Warn("Found metal sheet with unexpected identifier", "sheet", sheet.ID, "identifier", sheet.Identifier)
		}
	}

	return filteredSheets, nil
}

func (s *MetalSheets) Add(sheet *models.MetalSheet) (models.MetalSheetID, error) {
	slog.Debug("Adding metal sheet", "tool", sheet.ToolID, "identifier", sheet.Identifier)

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

	sheet.ID = models.MetalSheetID(id)
	return sheet.ID, nil
}

func (s *MetalSheets) Update(sheet *models.MetalSheet) error {
	slog.Debug("Updating metal sheet", "sheet", sheet.ID)

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

func (s *MetalSheets) AssignTool(sheetID models.MetalSheetID, toolID int64) error {
	slog.Debug("Assigning tool to metal sheet", "sheet", sheetID, "tool", toolID)

	if toolID <= 0 {
		return errors.NewValidationError("tool_id: must be positive")
	}

	if _, err := s.Get(sheetID); err != nil {
		return fmt.Errorf("get sheet for tool assignment: %v", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET tool_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2;
	`, TableNameMetalSheets)

	if _, err := s.DB.Exec(query, toolID, sheetID); err != nil {
		return s.GetUpdateError(err)
	}

	return nil
}

func (s *MetalSheets) Delete(id models.MetalSheetID) error {
	slog.Debug("Deleting metal sheet", "id", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, TableNameMetalSheets)
	if _, err := s.DB.Exec(query, id); err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func scanMetalSheet(scanner Scannable) (*models.MetalSheet, error) {
	sheet := &models.MetalSheet{}
	var identifierStr string

	err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight,
		&sheet.STF, &sheet.STFMax, &identifierStr, &sheet.ToolID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan metal sheet: %v", err)
	}

	sheet.Identifier = models.MachineType(identifierStr)
	return sheet, nil
}
