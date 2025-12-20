package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateUpperMetalSheet string = `
		INSERT INTO metal_sheets (tool_id, tile_height, value, type)
		VALUES (:tool_id, :tile_height, :value, 'upper');
	`
	SQLGetUpperMetalSheetByID string = `
		SELECT id, tool_id, tile_height, value
		FROM metal_sheets
		WHERE id = :id AND type = 'upper';
	`
	SQLUpdateUpperMetalSheet string = `
		UPDATE metal_sheets
		SET tool_id = :tool_id,
		    tile_height = :tile_height,
		    value = :value,
		    type = "upper"
		WHERE id = :id;
	`
	SQLDeleteUpperMetalSheet string = `
		DELETE FROM metal_sheets
		WHERE id = :id AND type = 'upper';
	`
	SQLListUpperMetalSheets string = `
		SELECT id, tool_id, tile_height, value
		FROM metal_sheets
		WHERE type = 'upper';
	`
)

type UpperMetalSheetsService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewUpperMetalSheetsService(c *shared.Config) *UpperMetalSheetsService {
	return &UpperMetalSheetsService{
		BaseService: shared.NewBaseService(c, "UpperMetalSheet"),
		mx:          &sync.Mutex{},
	}
}

func (s *UpperMetalSheetsService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateMetalSheetTable)
}

func (s *UpperMetalSheetsService) GetByID(id shared.EntityID) (*shared.UpperMetalSheet, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	var upperSheet shared.UpperMetalSheet
	r := s.DB().QueryRow(SQLGetUpperMetalSheetByID, sql.Named("id", id))
	err := r.Scan(
		&upperSheet.ID,
		&upperSheet.ToolID,
		&upperSheet.TileHeight,
		&upperSheet.Value,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewValidationError("upper metal sheet not found").MasterError()
		}
		return nil, errors.NewMasterError(err, 0)
	}

	return &upperSheet, nil
}

func (s *UpperMetalSheetsService) List() ([]*shared.UpperMetalSheet, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListUpperMetalSheets)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	upperSheets := []*shared.UpperMetalSheet{}
	for rows.Next() {
		var upperSheet shared.UpperMetalSheet
		if err := rows.Scan(
			&upperSheet.ID,
			&upperSheet.ToolID,
			&upperSheet.TileHeight,
			&upperSheet.Value,
		); err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		upperSheets = append(upperSheets, &upperSheet)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return upperSheets, nil
}

func (s *UpperMetalSheetsService) Create(entity *shared.UpperMetalSheet) (*shared.UpperMetalSheet, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	r, err := s.DB().Exec(SQLCreateUpperMetalSheet,
		sql.Named("tool_id", entity.ToolID),
		sql.Named("tile_height", entity.TileHeight),
		sql.Named("value", entity.Value),
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID returned after insert: %v", id).MasterError()
	}

	entity.ID = shared.EntityID(id)

	return entity, nil
}

func (s *UpperMetalSheetsService) Update(entity *shared.UpperMetalSheet) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	_, err := s.DB().Exec(SQLUpdateUpperMetalSheet,
		sql.Named("id", entity.ID),
		sql.Named("tool_id", entity.ToolID),
		sql.Named("tile_height", entity.TileHeight),
		sql.Named("value", entity.Value),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *UpperMetalSheetsService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteUpperMetalSheet,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

var _ shared.Service[*shared.UpperMetalSheet, shared.EntityID] = (*UpperMetalSheetsService)(nil)
