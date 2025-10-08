package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// MetalSheets represents a collection of metal sheets in the database.
type MetalSheets struct {
	*BaseService
	notes *Notes
}

func NewMetalSheets(db *sql.DB, notes *Notes) *MetalSheets {
	base := NewBaseService(db, "Metal Sheets")

	query := `
		CREATE TABLE IF NOT EXISTS metal_sheets (
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
	`

	if err := base.CreateTable(query, "metal_sheets"); err != nil {
		panic(err)
	}

	return &MetalSheets{
		BaseService: base,
		notes:       notes,
	}
}

// List returns all metal sheets
func (s *MetalSheets) List() ([]*models.MetalSheet, error) {
	s.LogOperation("Listing metal sheets")

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, s.HandleSelectError(err, "metal_sheets")
	}
	defer rows.Close()

	sheets, err := ScanMetalSheetsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Listed metal sheets", fmt.Sprintf("count: %d", len(sheets)))
	return sheets, nil
}

// Get returns a metal sheet by ID
func (s *MetalSheets) Get(id int64) (*models.MetalSheet, error) {
	if err := ValidateID(id, "metal_sheet"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting metal sheet", id)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE id = $1;
	`
	row := s.db.QueryRow(query, id)

	sheet, err := ScanSingleRow(row, ScanMetalSheet, "metal_sheets")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("metal sheet with ID %d", id))
		}
		return nil, err
	}

	return sheet, nil
}

// GetWithNotes returns a metal sheet with its related notes loaded
func (s *MetalSheets) GetWithNotes(id int64) (*models.MetalSheetWithNotes, error) {
	if err := ValidateID(id, "metal_sheet"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting metal sheet with notes", id)

	sheet, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	result := &models.MetalSheetWithNotes{
		MetalSheet:  sheet,
		LoadedNotes: []*models.Note{},
	}

	s.LogOperation("Found metal sheet with notes", fmt.Sprintf("id: %d", id))
	return result, nil
}

// GetByToolID returns all metal sheets assigned to a specific tool
func (s *MetalSheets) GetByToolID(toolID int64) ([]*models.MetalSheet, error) {
	if err := ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting metal sheets for tool", toolID)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE tool_id = $1
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query, toolID)
	if err != nil {
		return nil, s.HandleSelectError(err, "metal_sheets")
	}
	defer rows.Close()

	sheets, err := ScanMetalSheetsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Found metal sheets for tool", fmt.Sprintf("tool: %d, count: %d", toolID, len(sheets)))
	return sheets, nil
}

// GetByMachineType returns all metal sheets of the specified machine type
func (s *MetalSheets) GetByMachineType(machineType models.MachineType) ([]*models.MetalSheet, error) {
	if !machineType.IsValid() {
		return nil, utils.NewValidationError(fmt.Sprintf("invalid machine type: %s", machineType))
	}

	s.LogOperation("Getting metal sheets for machine type", string(machineType))

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE identifier = $1
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query, machineType.String())
	if err != nil {
		return nil, s.HandleSelectError(err, "metal_sheets")
	}
	defer rows.Close()

	sheets, err := ScanMetalSheetsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Found metal sheets for machine type", fmt.Sprintf("type: %s, count: %d", machineType, len(sheets)))
	return sheets, nil
}

// GetForPress returns metal sheets for tools on the specified press, filtered by the appropriate machine type
// Press 0 and 5 use SACMI machines, all others use SITI machines
func (s *MetalSheets) GetForPress(pressNumber models.PressNumber, toolsMap map[int64]*models.Tool) ([]*models.MetalSheet, error) {
	s.LogOperation("Getting metal sheets for press", fmt.Sprintf("press: %d, tools: %d", pressNumber, len(toolsMap)))

	// Get the expected machine type for this press
	expectedMachineType := models.GetMachineTypeForPress(pressNumber)
	s.LogOperation("Press machine type determined", fmt.Sprintf("press: %d, type: %s", pressNumber, expectedMachineType))

	// Get metal sheets for all tools on this press
	var allSheets models.MetalSheetList
	for toolID := range toolsMap {
		sheets, err := s.GetByToolID(toolID)
		if err != nil {
			s.log.Error("Failed to get metal sheets for tool %d: %v", toolID, err)
			continue
		}
		s.LogOperation("Retrieved sheets for tool", fmt.Sprintf("tool: %d, count: %d", toolID, len(sheets)))
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
			s.log.Warn("Found metal sheet %d with unexpected identifier: %s", sheet.ID, sheet.Identifier)
		}
	}

	s.LogOperation("Metal sheet distribution calculated",
		fmt.Sprintf("press: %d, SACMI: %d, SITI: %d, other: %d, total: %d, filtered: %d",
			pressNumber, sacmiCount, sitiCount, otherCount, len(allSheets), len(filteredSheets)))

	return filteredSheets, nil
}

// GetAvailable returns all metal sheets (tool_id is now required so all sheets are assigned)
// This method is kept for backward compatibility but now returns all sheets
func (s *MetalSheets) GetAvailable() ([]*models.MetalSheet, error) {
	s.LogOperation("Getting all metal sheets (backward compatibility)")
	return s.List()
}

// Add inserts a new metal sheet
func (s *MetalSheets) Add(sheet *models.MetalSheet) (int64, error) {
	if err := ValidateMetalSheet(sheet); err != nil {
		return 0, err
	}

	s.LogOperation("Adding metal sheet", fmt.Sprintf("tool_id: %d, identifier: %s", sheet.ToolID, sheet.Identifier))

	query := `
		INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, identifier, tool_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	result, err := s.db.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.Identifier.String(), sheet.ToolID)
	if err != nil {
		return 0, s.HandleInsertError(err, "metal_sheets")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.HandleInsertError(err, "metal_sheets")
	}

	// Set sheet ID for return
	sheet.ID = id

	s.LogOperation("Added metal sheet", fmt.Sprintf("id: %d", id))
	return id, nil
}

// Update updates an existing metal sheet
func (s *MetalSheets) Update(sheet *models.MetalSheet) error {
	if err := ValidateMetalSheet(sheet); err != nil {
		return err
	}

	if err := ValidateID(sheet.ID, "metal_sheet"); err != nil {
		return err
	}

	s.LogOperation("Updating metal sheet", fmt.Sprintf("id: %d", sheet.ID))

	query := `
		UPDATE metal_sheets
		SET tile_height = $1, value = $2, marke_height = $3, stf = $4, stf_max = $5,
			identifier = $6, tool_id = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8;
	`

	result, err := s.db.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.Identifier.String(), sheet.ToolID, sheet.ID)
	if err != nil {
		return s.HandleUpdateError(err, "metal_sheets")
	}

	if err := s.CheckRowsAffected(result, "metal_sheet", sheet.ID); err != nil {
		return err
	}

	s.LogOperation("Updated metal sheet", fmt.Sprintf("id: %d", sheet.ID))
	return nil
}

// AssignTool assigns a metal sheet to a tool
func (s *MetalSheets) AssignTool(sheetID int64, toolID int64) error {
	if err := ValidateID(sheetID, "metal_sheet"); err != nil {
		return err
	}

	if err := ValidateID(toolID, "tool"); err != nil {
		return err
	}

	s.LogOperation("Assigning tool to metal sheet", fmt.Sprintf("sheet_id: %d, tool_id: %d", sheetID, toolID))

	// Get current sheet to verify it exists
	_, err := s.Get(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get sheet for tool assignment: %v", err)
	}

	// Update the tool assignment
	result, err := s.db.Exec(
		`UPDATE metal_sheets
		SET tool_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`,
		toolID, sheetID,
	)
	if err != nil {
		return s.HandleUpdateError(err, "metal_sheets")
	}

	if err := s.CheckRowsAffected(result, "metal_sheet", sheetID); err != nil {
		return err
	}

	s.LogOperation("Assigned tool to metal sheet", fmt.Sprintf("sheet_id: %d, tool_id: %d", sheetID, toolID))
	return nil
}

// Delete deletes a metal sheet
func (s *MetalSheets) Delete(id int64) error {
	if err := ValidateID(id, "metal_sheet"); err != nil {
		return err
	}

	s.LogOperation("Deleting metal sheet", id)

	query := `DELETE FROM metal_sheets WHERE id = $1;`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return s.HandleDeleteError(err, "metal_sheets")
	}

	if err := s.CheckRowsAffected(result, "metal_sheet", id); err != nil {
		return err
	}

	s.LogOperation("Deleted metal sheet", id)
	return nil
}
