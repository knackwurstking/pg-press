package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// MetalSheets represents a collection of metal sheets in the database.
type MetalSheets struct {
	db    *sql.DB
	notes *Notes
	log   *logger.Logger
}

func NewMetalSheets(db *sql.DB, notes *Notes) *MetalSheets {
	metalSheet := &MetalSheets{
		db:    db,
		notes: notes,
		log:   logger.GetComponentLogger("Service: Metal Sheets"),
	}

	if err := metalSheet.createTable(); err != nil {
		panic(err)
	}

	return metalSheet
}

func (s *MetalSheets) createTable() error {
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

	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create metal_sheets table: %v", err)
	}

	return nil
}

// List returns all metal sheets
func (s *MetalSheets) List() ([]*models.MetalSheet, error) {
	s.log.Info("Listing metal sheets")

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet

	for rows.Next() {
		sheet, err := s.scanMetalSheet(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metal sheet: %v", err)
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}

	s.log.Debug("Listed %d metal sheets", len(sheets))
	return sheets, nil
}

// Get returns a metal sheet by ID
func (s *MetalSheets) Get(id int64) (*models.MetalSheet, error) {
	s.log.Info("Getting metal sheet, id: %d", id)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE id = $1;
	`
	row := s.db.QueryRow(query, id)

	sheet, err := s.scanMetalSheet(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("metal sheet with ID %d", id))
		}
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}

	return sheet, nil
}

// GetWithNotes returns a metal sheet with its related notes loaded
func (s *MetalSheets) GetWithNotes(id int64) (*models.MetalSheetWithNotes, error) {
	s.log.Info("Getting metal sheet with notes, id: %d", id)

	sheet, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	result := &models.MetalSheetWithNotes{
		MetalSheet:  sheet,
		LoadedNotes: []*models.Note{},
	}

	return result, nil
}

// GetByToolID returns all metal sheets assigned to a specific tool
func (s *MetalSheets) GetByToolID(toolID int64) ([]*models.MetalSheet, error) {
	s.log.Info("Getting metal sheets for tool, id: %d", toolID)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE tool_id = $1
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet

	for rows.Next() {
		sheet, err := s.scanMetalSheet(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metal sheet: %v", err)
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}

	s.log.Debug("Found %d metal sheets for tool %d", len(sheets), toolID)
	return sheets, nil
}

// GetByMachineType returns all metal sheets of the specified machine type
func (s *MetalSheets) GetByMachineType(machineType models.MachineType) ([]*models.MetalSheet, error) {
	s.log.Info("Getting metal sheets for machine type: %s", machineType)

	if !machineType.IsValid() {
		return nil, fmt.Errorf("invalid machine type: %s", machineType)
	}

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM metal_sheets
		WHERE identifier = $1
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query, machineType.String())
	if err != nil {
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet

	for rows.Next() {
		sheet, err := s.scanMetalSheet(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metal sheet: %v", err)
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}

	s.log.Debug("Found %d metal sheets for machine type %s", len(sheets), machineType)
	return sheets, nil
}

// GetForPress returns metal sheets for tools on the specified press, filtered by the appropriate machine type
// Press 0 and 5 use SACMI machines, all others use SITI machines
func (s *MetalSheets) GetForPress(pressNumber models.PressNumber, toolsMap map[int64]*models.Tool) ([]*models.MetalSheet, error) {
	s.log.Info("Getting metal sheets for press %d with machine type filtering", pressNumber)

	// Get the expected machine type for this press
	expectedMachineType := models.GetMachineTypeForPress(pressNumber)
	s.log.Debug("Press %d should use %s machines", pressNumber, expectedMachineType)

	// Get metal sheets for all tools on this press
	var allSheets models.MetalSheetList
	for toolID := range toolsMap {
		sheets, err := s.GetByToolID(toolID)
		if err != nil {
			s.log.Error("Failed to get metal sheets for tool %d: %v", toolID, err)
			continue
		}
		s.log.Debug("Tool %d has %d metal sheets", toolID, len(sheets))
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

	s.log.Debug("Metal sheet distribution for press %d: %d SACMI, %d SITI, %d other (total: %d)",
		pressNumber, sacmiCount, sitiCount, otherCount, len(allSheets))
	s.log.Debug("Filtered to %d %s metal sheets for press %d",
		len(filteredSheets), expectedMachineType, pressNumber)
	return filteredSheets, nil
}

// GetAvailable returns all metal sheets (tool_id is now required so all sheets are assigned)
// This method is kept for backward compatibility but now returns all sheets
func (s *MetalSheets) GetAvailable() ([]*models.MetalSheet, error) {
	s.log.Info("Getting all metal sheets (tool_id is now required)")
	return s.List()
}

// Add inserts a new metal sheet
func (s *MetalSheets) Add(sheet *models.MetalSheet) (int64, error) {
	s.log.Info("Adding metal sheet: %s", sheet.String())

	query := `
		INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, identifier, tool_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	result, err := s.db.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.Identifier.String(), sheet.ToolID)
	if err != nil {
		return 0, fmt.Errorf("insert error: metal_sheets: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert error: metal_sheets: %v", err)
	}

	// Set sheet ID for return
	sheet.ID = id

	s.log.Debug("Added metal sheet with ID %d", id)

	return id, nil
}

// Update updates an existing metal sheet
func (s *MetalSheets) Update(sheet *models.MetalSheet) error {
	s.log.Info("Updating metal sheet: %d", sheet.ID)

	query := `
		UPDATE metal_sheets
		SET tile_height = $1, value = $2, marke_height = $3, stf = $4, stf_max = $5,
						identifier = $6, tool_id = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8;
	`

	_, err := s.db.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.Identifier.String(), sheet.ToolID, sheet.ID)
	if err != nil {
		return fmt.Errorf("update error: metal_sheets: %v", err)
	}

	s.log.Debug("Updated metal sheet with ID %d", sheet.ID)

	return nil
}

// AssignTool assigns a metal sheet to a tool
func (s *MetalSheets) AssignTool(sheetID int64, toolID int64) error {
	s.log.Info("Assigning tool to metal sheet: sheet_id=%d, tool_id=%v", sheetID, toolID)

	// Get current sheet to track changes
	sheet, err := s.Get(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get sheet for tool assignment: %v", err)
	}

	// Update the tool assignment
	sheet.ToolID = toolID

	// Update both tool_id and mods in database
	_, err = s.db.Exec(
		`UPDATE metal_sheets
		SET tool_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`,
		toolID, sheetID,
	)
	if err != nil {
		return fmt.Errorf("update error: metal_sheets: %v", err)
	}

	s.log.Debug("Assigned tool %v to metal sheet %d", toolID, sheetID)

	return nil
}

// Delete deletes a metal sheet
func (s *MetalSheets) Delete(id int64) error {
	s.log.Info("Deleting metal sheet: %d", id)

	query := `
		DELETE FROM metal_sheets WHERE id = $1;
	`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete error: metal_sheets: %v", err)
	}

	s.log.Debug("Deleted metal sheet with ID %d", id)

	return nil
}

func (s *MetalSheets) scanMetalSheet(scanner interfaces.Scannable) (*models.MetalSheet, error) {
	sheet := &models.MetalSheet{}

	var (
		toolID int64
	)

	var identifierStr string
	if err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight, &sheet.STF, &sheet.STFMax,
		&identifierStr, &toolID); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan error: metal_sheets: %v", err)
	}

	// tool_id is now required
	sheet.ToolID = toolID

	// Convert string identifier to MachineType
	sheet.Identifier = models.MachineType(identifierStr)

	return sheet, nil
}
