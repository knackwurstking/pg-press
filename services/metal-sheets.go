// Service for managing metal sheets with validation
package services

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

type MetalSheets struct {
	*Base
}

func NewMetalSheets(r *Registry) *MetalSheets {
	return &MetalSheets{
		Base: NewBase(r),
	}
}

func (s *MetalSheets) Get(id models.MetalSheetID) (*models.MetalSheet, *errors.MasterError) {
	row := s.DB.QueryRow(SQLGetMetalSheet, id)

	sheet, err := ScanMetalSheet(row)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return sheet, nil
}

func (s *MetalSheets) List() ([]*models.MetalSheet, *errors.MasterError) {
	rows, err := s.DB.Query(SQLListMetalSheets)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanMetalSheet)
}

func (s *MetalSheets) ListByToolID(toolID models.ToolID) ([]*models.MetalSheet, *errors.MasterError) {
	rows, err := s.DB.Query(SQLListMetalSheetsByToolID, toolID)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanMetalSheet)
}

func (s *MetalSheets) ListByMachineType(machineType models.MachineType) ([]*models.MetalSheet, *errors.MasterError) {
	rows, err := s.DB.Query(SQLListMetalSheetsByMachineType, machineType.String())
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

// TODO: Create `AddUpperMetalSheet` method
//func (s *MetalSheets) AddUpperMetalSheet(...) (models.MetalSheetID, *errors.MasterError) {}

// TODO: Create `AddLowerMetalSheet` method
//func (s *MetalSheets) AddLowerMetalSheet(...) (models.MetalSheetID, *errors.MasterError) {}

// TODO: Remove this method once specific methods for upper and lower sheets are implemented
func (s *MetalSheets) Add(sheet *models.MetalSheet) (models.MetalSheetID, *errors.MasterError) {
	verr := sheet.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	result, err := s.DB.Exec(
		SQLAddMetalSheet,
		sheet.TileHeight,
		sheet.Value,
		sheet.MarkeHeight,
		sheet.STF,
		sheet.STFMax,
		sheet.Identifier.String(),
		sheet.ToolID,
		sheet.UpdatedAt,
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

// TODO: Create `UpdateUpperMetalSheet` method
//func (s *MetalSheets) UpdateUpperMetalSheet(...) *errors.MasterError {}

// TODO: Create `UpdateLowerMetalSheet` method
//func (s *MetalSheets) UpdateLowerMetalSheet(...) *errors.MasterError {}

// TODO: Remove this method once specific methods for upper and lower sheets are implemented
func (s *MetalSheets) Update(sheet *models.MetalSheet) *errors.MasterError {
	verr := sheet.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	_, err := s.DB.Exec(SQLUpdateMetalSheet,
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

	_, err := s.DB.Exec(SQLUpdateMetalSheetToolID, toolID, sheetID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *MetalSheets) Delete(id models.MetalSheetID) *errors.MasterError {
	_, err := s.DB.Exec(SQLDeleteMetalSheet, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
