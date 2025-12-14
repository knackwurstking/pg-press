package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateMetalSheetTable string = `
		CREATE TABLE IF NOT EXISTS metal_sheets (
			id 				INTEGER NOT NULL,
			tool_id 		INTEGER NOT NULL,
			tile_height 	REAL NOT NULL,
			value 			REAL NOT NULL,
			type 			TEXT NOT NULL,
			marke_height 	INTEGER,
			stf 			REAL,
			stf_max 		REAL,
			identifier 		TEXT,

			PRIMARY KEY("id" AUTOINCREMENT),
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
	`
	SQLCreateMetalSheet string = `
		INSERT INTO metal_sheets (tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier)
		VALUES (:tool_id, :tile_height, :value, :type, :marke_height, :stf, :stf_max, :identifier);
	`
	SQLGetMetalSheetByID string = `
		SELECT id, tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier
		FROM metal_sheets
		WHERE id = :id;
	`
	SQLUpdateMetalSheet string = `
		UPDATE metal_sheets
		SET tool_id 		= :tool_id,
			tile_height 	= :tile_height,
			value 			= :value,
			type 			= :type,
			marke_height 	= :marke_height,
			stf 			= :stf,
			stf_max 		= :stf_max,
			identifier 		= :identifier
		WHERE id = :id;
	`
	SQLDeleteMetalSheet string = `
		DELETE FROM metal_sheets
		WHERE id = :id;
	`
	SQLListMetalSheets string = `
		SELECT id, tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier
		FROM metal_sheets;
	`
)

type MetalSheetService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewMetalSheetService(c *shared.Config) *MetalSheetService {
	return &MetalSheetService{
		BaseService: &shared.BaseService{
			Config: c,
		},

		mx: &sync.Mutex{},
	}
}

func (s *MetalSheetService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateMetalSheetTable)
}

func (s *MetalSheetService) Create(entity any) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	var sqlCreate string
	var params []any

	switch e := entity.(type) {
	case *shared.LowerMetalSheet:
		verr := e.Validate()
		if verr != nil {
			return verr.MasterError()
		}
		sqlCreate = SQLCreateMetalSheet
		params = []any{
			sql.Named("tool_id", e.ToolID),
			sql.Named("tile_height", e.TileHeight),
			sql.Named("value", e.Value),
			sql.Named("type", "lower"),
			sql.Named("marke_height", e.MarkeHeight),
			sql.Named("stf", e.STF),
			sql.Named("stf_max", e.STFMax),
			sql.Named("identifier", e.Identifier),
		}
	case *shared.UpperMetalSheet:
		verr := e.Validate()
		if verr != nil {
			return verr.MasterError()
		}
		sqlCreate = SQLCreateMetalSheet
		params = []any{
			sql.Named("tool_id", e.ToolID),
			sql.Named("tile_height", e.TileHeight),
			sql.Named("value", e.Value),
			sql.Named("type", "upper"),
			sql.Named("marke_height", nil),
			sql.Named("stf", nil),
			sql.Named("stf_max", nil),
			sql.Named("identifier", nil),
		}
	default:
		return errors.NewValidationError("invalid entity type").MasterError()
	}

	r, err := s.DB().Exec(sqlCreate, params...)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	if id <= 0 {
		return errors.NewMasterError(
			errors.NewValidationError("invalid ID returned after insert: %v", id), 0)
	}

	switch e := entity.(type) {
	case *shared.LowerMetalSheet:
		e.ID = shared.EntityID(id)
	case *shared.UpperMetalSheet:
		e.ID = shared.EntityID(id)
	}

	return nil
}

func (s *MetalSheetService) Update(entity any) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	var sqlUpdate string
	var params []any

	switch e := entity.(type) {
	case *shared.LowerMetalSheet:
		verr := e.Validate()
		if verr != nil {
			return verr.MasterError()
		}
		sqlUpdate = SQLUpdateMetalSheet
		params = []any{
			sql.Named("id", e.ID),
			sql.Named("tool_id", e.ToolID),
			sql.Named("tile_height", e.TileHeight),
			sql.Named("value", e.Value),
			sql.Named("type", "lower"),
			sql.Named("marke_height", e.MarkeHeight),
			sql.Named("stf", e.STF),
			sql.Named("stf_max", e.STFMax),
			sql.Named("identifier", e.Identifier),
		}
	case *shared.UpperMetalSheet:
		verr := e.Validate()
		if verr != nil {
			return verr.MasterError()
		}
		sqlUpdate = SQLUpdateMetalSheet
		params = []any{
			sql.Named("id", e.ID),
			sql.Named("tool_id", e.ToolID),
			sql.Named("tile_height", e.TileHeight),
			sql.Named("value", e.Value),
			sql.Named("type", "upper"),
			sql.Named("marke_height", nil),
			sql.Named("stf", nil),
			sql.Named("stf_max", nil),
			sql.Named("identifier", nil),
		}
	default:
		return errors.NewValidationError("invalid entity type").MasterError()
	}

	_, err := s.DB().Exec(sqlUpdate, params...)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *MetalSheetService) GetByID(id shared.EntityID) (any, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetMetalSheetByID,
		sql.Named("id", id),
	)

	var metalSheetType string
	var metalSheetID shared.EntityID
	var toolID shared.EntityID
	var tileHeight float64
	var value float64

	err := r.Scan(
		&metalSheetID,
		&toolID,
		&tileHeight,
		&value,
		&metalSheetType,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	if metalSheetType == "lower" {
		var markeHeight int
		var stf float64
		var stfMax float64
		var identifier string

		err := r.Scan(
			&markeHeight,
			&stf,
			&stfMax,
			&identifier,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}

		lowerSheet := &shared.LowerMetalSheet{
			BaseMetalSheet: shared.BaseMetalSheet{
				ID:         metalSheetID,
				ToolID:     toolID,
				TileHeight: tileHeight,
				Value:      value,
			},
			MarkeHeight: markeHeight,
			STF:         stf,
			STFMax:      stfMax,
			Identifier:  shared.MachineType(identifier),
		}

		return lowerSheet, nil
	} else {
		upperSheet := &shared.UpperMetalSheet{
			BaseMetalSheet: shared.BaseMetalSheet{
				ID:         metalSheetID,
				ToolID:     toolID,
				TileHeight: tileHeight,
				Value:      value,
			},
		}

		return upperSheet, nil
	}
}

func (s *MetalSheetService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteMetalSheet,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *MetalSheetService) List() ([]any, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListMetalSheets)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	metalSheets := []any{}
	for rows.Next() {
		var metalSheetType string
		var metalSheetID shared.EntityID
		var toolID shared.EntityID
		var tileHeight float64
		var value float64

		err := rows.Scan(
			&metalSheetID,
			&toolID,
			&tileHeight,
			&value,
			&metalSheetType,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}

		if metalSheetType == "lower" {
			var markeHeight int
			var stf float64
			var stfMax float64
			var identifier string

			err := rows.Scan(
				&markeHeight,
				&stf,
				&stfMax,
				&identifier,
			)
			if err != nil {
				return nil, errors.NewMasterError(err, 0)
			}

			lowerSheet := &shared.LowerMetalSheet{
				BaseMetalSheet: shared.BaseMetalSheet{
					ID:         metalSheetID,
					ToolID:     toolID,
					TileHeight: tileHeight,
					Value:      value,
				},
				MarkeHeight: markeHeight,
				STF:         stf,
				STFMax:      stfMax,
				Identifier:  shared.MachineType(identifier),
			}

			metalSheets = append(metalSheets, lowerSheet)
		} else {
			upperSheet := &shared.UpperMetalSheet{
				BaseMetalSheet: shared.BaseMetalSheet{
					ID:         metalSheetID,
					ToolID:     toolID,
					TileHeight: tileHeight,
					Value:      value,
				},
			}

			metalSheets = append(metalSheets, upperSheet)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return metalSheets, nil
}
