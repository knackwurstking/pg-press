package metalsheets

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/entities/notes"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService

	notes *notes.Service
}

func NewService(db *sql.DB, notes *notes.Service) *Service {
	base := base.NewBaseService(db, "Metal Sheets")

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

	return &Service{
		BaseService: base,
		notes:       notes,
	}
}

// List returns all metal sheets
func (s *Service) List() ([]*models.MetalSheet, error) {
	s.Log.Debug("Listing metal sheets")

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		ORDER BY id DESC;
	`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, s.HandleSelectError(err, "metal_sheets")
	}
	defer rows.Close()

	sheets, err := scanMetalSheetsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Debug("Listed metal sheets: count: %d", len(sheets))
	return sheets, nil
}

func (s *Service) Get(id int64) (*models.MetalSheet, error) {
	if err := validation.ValidateID(id, "metal_sheet"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting metal sheet: %d", id)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE id = $1;
	`
	row := s.DB.QueryRow(query, id)

	sheet, err := scanner.ScanSingleRow(row, scanMetalSheet, "metal_sheets")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("metal sheet with ID %d", id))
		}
		return nil, err
	}

	return sheet, nil
}

func (s *Service) GetWithNotes(id int64) (*models.MetalSheetWithNotes, error) {
	if err := validation.ValidateID(id, "metal_sheet"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting metal sheet with notes: %d", id)

	sheet, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	result := &models.MetalSheetWithNotes{
		MetalSheet:  sheet,
		LoadedNotes: []*models.Note{},
	}

	s.Log.Debug("Found metal sheet with notes: id: %d", id)
	return result, nil
}

// GetByToolID returns all metal sheets assigned to a specific tool
func (s *Service) GetByToolID(toolID int64) ([]*models.MetalSheet, error) {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting metal sheets for tool: %d", toolID)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE tool_id = $1
		ORDER BY id DESC;
	`
	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, s.HandleSelectError(err, "metal_sheets")
	}
	defer rows.Close()

	sheets, err := scanMetalSheetsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Debug("Found metal sheets for tool: tool: %d, count: %d", toolID, len(sheets))
	return sheets, nil
}

func (s *Service) GetByMachineType(machineType models.MachineType) ([]*models.MetalSheet, error) {
	if !machineType.IsValid() {
		return nil, utils.NewValidationError(fmt.Sprintf("invalid machine type: %s", machineType))
	}

	s.Log.Debug("Getting metal sheets for machine type: %s", string(machineType))

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE identifier = $1
		ORDER BY id DESC;
	`
	rows, err := s.DB.Query(query, machineType.String())
	if err != nil {
		return nil, s.HandleSelectError(err, "metal_sheets")
	}
	defer rows.Close()

	sheets, err := scanMetalSheetsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Debug("Found metal sheets for machine type: type: %s, count: %d",
		machineType, len(sheets))
	return sheets, nil
}

func (s *Service) GetForPress(pressNumber models.PressNumber, toolsMap map[int64]*models.Tool) ([]*models.MetalSheet, error) {
	s.Log.Debug("Getting metal sheets for press: %d, tools: %d",
		pressNumber, len(toolsMap))

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

	s.Log.Debug(
		"Metal sheet distribution calculated: "+
			"press: %d, SACMI: %d, SITI: %d, other: %d, total: %d, filtered: %d",
		pressNumber, sacmiCount, sitiCount, otherCount, len(allSheets), len(filteredSheets),
	)

	return filteredSheets, nil
}

func (s *Service) GetAvailable() ([]*models.MetalSheet, error) {
	s.Log.Debug("Getting all metal sheets (backward compatibility)")
	return s.List()
}

func (s *Service) Add(sheet *models.MetalSheet) (int64, error) {
	s.Log.Debug("Adding metal sheet: tool_id: %d, identifier: %s",
		sheet.ToolID, sheet.Identifier)

	query := `
		INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, identifier, tool_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

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
		return 0, s.HandleInsertError(err, "metal_sheets")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.HandleInsertError(err, "metal_sheets")
	}

	// Set sheet ID for return
	sheet.ID = id

	s.Log.Debug("Added metal sheet: %d", id)
	return id, nil
}

func (s *Service) Update(sheet *models.MetalSheet) error {
	if err := validation.ValidateID(sheet.ID, "metal_sheet"); err != nil {
		return err
	}

	s.Log.Debug("Updating metal sheet: %d", sheet.ID)

	query := `
		UPDATE metal_sheets
		SET tile_height = $1, value = $2, marke_height = $3, stf = $4, stf_max = $5,
			identifier = $6, tool_id = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8;
	`

	result, err := s.DB.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.Identifier.String(), sheet.ToolID, sheet.ID)
	if err != nil {
		return s.HandleUpdateError(err, "metal_sheets")
	}

	if err := s.CheckRowsAffected(result, "metal_sheet", sheet.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) AssignTool(sheetID int64, toolID int64) error {
	if err := validation.ValidateID(sheetID, "metal_sheet"); err != nil {
		return err
	}

	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	s.Log.Debug("Assigning tool to metal sheet: sheet_id: %d, tool_id: %d", sheetID, toolID)

	// Get current sheet to verify it exists
	_, err := s.Get(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get sheet for tool assignment: %v", err)
	}

	// Update the tool assignment
	result, err := s.DB.Exec(
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

	s.Log.Debug("Assigned tool to metal sheet: sheet_id: %d, tool_id: %d", sheetID, toolID)
	return nil
}

func (s *Service) Delete(id int64) error {
	if err := validation.ValidateID(id, "metal_sheet"); err != nil {
		return err
	}

	s.Log.Debug("Deleting metal sheet: %d", id)

	query := `DELETE FROM metal_sheets WHERE id = $1;`
	result, err := s.DB.Exec(query, id)
	if err != nil {
		return s.HandleDeleteError(err, "metal_sheets")
	}

	if err := s.CheckRowsAffected(result, "metal_sheet", id); err != nil {
		return err
	}

	s.Log.Debug("Deleted metal sheet: %d", id)
	return nil
}
