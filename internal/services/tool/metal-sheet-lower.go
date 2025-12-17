package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateLowerMetalSheet string = `
		INSERT INTO metal_sheets (tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier)
		VALUES (:tool_id, :tile_height, :value, 'lower', :marke_height, :stf, :stf_max, :identifier);
	`
	SQLGetLowerMetalSheetByID string = `
		SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
		FROM metal_sheets
		WHERE id = :id AND type = 'lower';
	`
	SQLUpdateLowerMetalSheet string = `
		UPDATE metal_sheets
		SET tool_id = :tool_id,
		    tile_height = :tile_height,
		    value = :value,
		    type = 'lower',
		    marke_height = :marke_height,
		    stf = :stf,
		    stf_max = :stf_max,
		    identifier = :identifier
		WHERE id = :id;
	`
	SQLDeleteLowerMetalSheet string = `
		DELETE FROM metal_sheets
		WHERE id = :id AND type = 'lower';
	`
	SQLListLowerMetalSheets string = `
		SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
		FROM metal_sheets
		WHERE type = 'lower';
	`
)

type LowerMetalSheetService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewLowerMetalSheetService(c *shared.Config) *LowerMetalSheetService {
	return &LowerMetalSheetService{
		BaseService: shared.NewBaseService(c, "LowerMetalSheet"),
		mx:          &sync.Mutex{},
	}
}

func (s *LowerMetalSheetService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateMetalSheetTable)
}

func (s *LowerMetalSheetService) GetByID(id shared.EntityID) (*shared.LowerMetalSheet, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	var lowerSheet shared.LowerMetalSheet
	r := s.DB().QueryRow(SQLGetLowerMetalSheetByID, sql.Named("id", id))
	err := r.Scan(
		&lowerSheet.ID,
		&lowerSheet.ToolID,
		&lowerSheet.TileHeight,
		&lowerSheet.Value,
		&lowerSheet.MarkeHeight,
		&lowerSheet.STF,
		&lowerSheet.STFMax,
		&lowerSheet.Identifier,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewValidationError("lower metal sheet not found").MasterError()
		}
		return nil, errors.NewMasterError(err, 0)
	}

	return &lowerSheet, nil
}

func (s *LowerMetalSheetService) List() ([]*shared.LowerMetalSheet, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListLowerMetalSheets)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	lowerSheets := []*shared.LowerMetalSheet{}
	for rows.Next() {
		var lowerSheet shared.LowerMetalSheet
		if err := rows.Scan(
			&lowerSheet.ID,
			&lowerSheet.ToolID,
			&lowerSheet.TileHeight,
			&lowerSheet.Value,
			&lowerSheet.MarkeHeight,
			&lowerSheet.STF,
			&lowerSheet.STFMax,
			&lowerSheet.Identifier,
		); err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		lowerSheets = append(lowerSheets, &lowerSheet)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return lowerSheets, nil
}

func (s *LowerMetalSheetService) Create(entity *shared.LowerMetalSheet) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	r, err := s.DB().Exec(SQLCreateLowerMetalSheet,
		sql.Named("tool_id", entity.ToolID),
		sql.Named("tile_height", entity.TileHeight),
		sql.Named("value", entity.Value),
		sql.Named("marke_height", entity.MarkeHeight),
		sql.Named("stf", entity.STF),
		sql.Named("stf_max", entity.STFMax),
		sql.Named("identifier", entity.Identifier),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	if id <= 0 {
		return errors.NewValidationError("invalid ID returned after insert: %v", id).MasterError()
	}

	entity.ID = shared.EntityID(id)

	return nil
}

func (s *LowerMetalSheetService) Update(entity *shared.LowerMetalSheet) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	_, err := s.DB().Exec(SQLUpdateLowerMetalSheet,
		sql.Named("id", entity.ID),
		sql.Named("tool_id", entity.ToolID),
		sql.Named("tile_height", entity.TileHeight),
		sql.Named("value", entity.Value),
		sql.Named("marke_height", entity.MarkeHeight),
		sql.Named("stf", entity.STF),
		sql.Named("stf_max", entity.STFMax),
		sql.Named("identifier", entity.Identifier),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *LowerMetalSheetService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteLowerMetalSheet,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

var _ shared.Service[*shared.LowerMetalSheet, shared.EntityID] = (*LowerMetalSheetService)(nil)
