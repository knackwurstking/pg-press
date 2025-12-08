package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

var ModificationTypes = []models.ModificationType{
	models.ModificationTypeTroubleReport,
	models.ModificationTypeMetalSheet,
	models.ModificationTypeTool,
	models.ModificationTypePressCycle,
	models.ModificationTypeUser,
	models.ModificationTypeNote,
	models.ModificationTypeAttachment,
}

type Modifications struct {
	*Base
}

func NewModifications(r *Registry) *Modifications {
	return &Modifications{
		Base: NewBase(r),
	}
}

func (s *Modifications) Add(
	mt models.ModificationType, mtID int64, data any, user models.TelegramID,
) (models.ModificationID, *errors.MasterError) {

	verr := s.validateModificationType(mt, mtID)
	if verr != nil {
		return 0, verr.MasterError()
	}

	if data == nil {
		return 0, errors.NewMasterError(fmt.Errorf("missing data"), http.StatusBadRequest)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, errors.NewMasterError(fmt.Errorf("marshal data: %#v", err), http.StatusBadRequest)
	}

	result, err := s.DB.Exec(SQLAddModification, user, mt, mtID, jsonData, time.Now())
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.ModificationID(id), nil
}

func (s *Modifications) Get(id models.ModificationID) (*models.Modification[any], *errors.MasterError) {
	row := s.DB.QueryRow(SQLGetModification, id)
	m, err := ScanModification(row)
	if err != nil {
		return m, errors.NewMasterError(err, 0)
	}
	return m, nil
}

func (s *Modifications) List(
	mt models.ModificationType, mtID int64, limit, offset int,
) ([]*models.Modification[any], *errors.MasterError) {

	verr := s.validateModificationType(mt, mtID)
	if verr != nil {
		return nil, verr.MasterError()
	}

	rows, err := s.DB.Query(SQLListModificationsByEntityType, mt, mtID, limit, offset)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) ListAll(mt models.ModificationType, mtID int64) ([]*models.Modification[any], *errors.MasterError) {
	return s.List(mt, mtID, -1, 0)
}

func (s *Modifications) Count(mt models.ModificationType, mtID int64) (int64, *errors.MasterError) {
	verr := s.validateModificationType(mt, mtID)
	if verr != nil {
		return 0, verr.MasterError()
	}

	var count int64
	err := s.DB.QueryRow(SQLCountModificationsByEntityType, mt, mtID).Scan(&count)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return count, nil
}

func (s *Modifications) Delete(id models.ModificationID) *errors.MasterError {
	_, err := s.DB.Exec(SQLDeleteModification, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *Modifications) DeleteAll(mt models.ModificationType, mtID int64) *errors.MasterError {
	verr := s.validateModificationType(mt, mtID)
	if verr != nil {
		return verr.MasterError()
	}

	_, err := s.DB.Exec(SQLDeleteModificationsByEntityType, mt, mtID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *Modifications) validateModificationType(mt models.ModificationType, mtID int64) *errors.ValidationError {
	if !slices.Contains(ModificationTypes, mt) {
		return errors.NewValidationError("invalid modification type: %#v", mt)
	}

	if mtID <= 0 {
		return errors.NewValidationError("modidfication type id cannot be lower or equal 0")
	}

	return nil
}
