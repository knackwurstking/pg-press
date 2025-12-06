// Service for managing metal sheets with validation
package services

import (
	"fmt"
	"net/http"

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

func (s *MetalSheets) Get(id models.MetalSheetID) (*models.MetalSheet, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		WHERE id = ?;
	`, TableNameMetalSheets)

	row := s.DB.QueryRow(query, id)

	sheet, err := ScanMetalSheet(row)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return sheet, nil
}

func (s *MetalSheets) List() ([]*models.MetalSheet, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		ORDER BY id DESC;
	`, TableNameMetalSheets)

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanMetalSheet)
}

func (s *MetalSheets) ListByToolID(toolID models.ToolID) ([]*models.MetalSheet, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		WHERE tool_id = $1
		ORDER BY id DESC;
	`, TableNameMetalSheets)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanMetalSheet)
}

func (s *MetalSheets) ListByMachineType(machineType models.MachineType) ([]*models.MetalSheet, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id
		FROM %s
		WHERE identifier = $1
		ORDER BY id DESC;
	`, TableNameMetalSheets)

	rows, err := s.DB.Query(query, machineType.String())
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanMetalSheet)
}

func (s *MetalSheets) ListByPress(pressNumber models.PressNumber, toolsMap map[models.ToolID]*models.Tool) ([]*models.MetalSheet, *errors.MasterError) {
	expectedMachineType := models.GetMachineTypeForPress(pressNumber)

	var allSheets models.MetalSheetList
	for toolID := range toolsMap {
		sheets, err := s.ListByToolID(toolID)
		if err != nil {
			return nil, err
		}
		allSheets = append(allSheets, sheets...)
	}

	var filteredSheets models.MetalSheetList
	for _, sheet := range allSheets {
		if sheet.Identifier == expectedMachineType {
			filteredSheets = append(filteredSheets, sheet)
		}
	}

	return filteredSheets, nil
}

func (s *MetalSheets) Add(sheet *models.MetalSheet) (models.MetalSheetID, *errors.MasterError) {
	if !sheet.Validate() {
		return 0, errors.NewMasterError(
			fmt.Errorf("invalid metal sheet data: %s", sheet),
			http.StatusBadRequest,
		)
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
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	sheet.ID = models.MetalSheetID(id)
	return sheet.ID, nil
}

func (s *MetalSheets) Update(sheet *models.MetalSheet) *errors.MasterError {
	if !sheet.Validate() {
		return errors.NewMasterError(
			fmt.Errorf("invalid metal sheet data: %s", sheet),
			http.StatusBadRequest,
		)
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
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *MetalSheets) AssignTool(sheetID models.MetalSheetID, toolID int64) *errors.MasterError {
	if toolID <= 0 {
		return errors.NewMasterError(
			fmt.Errorf("invalid tool id: %d", toolID),
			http.StatusBadRequest,
		)
	}

	_, merr := s.Get(sheetID)
	if merr != nil {
		return merr
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET tool_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2;
	`, TableNameMetalSheets)

	_, err := s.DB.Exec(query, toolID, sheetID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *MetalSheets) Delete(id models.MetalSheetID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, TableNameMetalSheets)

	if _, err := s.DB.Exec(query, id); err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
