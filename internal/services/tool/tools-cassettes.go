package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateCassette string = `
		INSERT INTO tools (position, width, height, type, code, cycles_offset, cycles, is_dead, min_thickness, max_thickness, model_type)
		VALUES (:position, :width, :height, :type, :code, :cycles_offset, :cycles, :is_dead, :min_thickness, :max_thickness, 'cassette');
	`

	SQLGetCassetteByID string = `
		SELECT id, position, width, height, type, code, cycles_offset, cycles, is_dead, min_thickness, max_thickness
		FROM tools
		WHERE id = :id AND model_type = 'cassette';
	`

	SQLUpdateCassette string = `
		UPDATE tools
		SET position = :position,
			width = :width,
			height = :height,
			type = :type,
			code = :code,
			cycles_offset = :cycles_offset,
			cycles = :cycles,
			is_dead = :is_dead,
			min_thickness = :min_thickness,
			max_thickness = :max_thickness
		WHERE id = :id AND model_type = 'cassette';
	`

	SQLDeleteCassette string = `
		DELETE FROM tools
		WHERE id = :id AND model_type = 'cassette';
	`

	SQLListCassettes string = `
		SELECT id, position, width, height, type, code, cycles_offset, cycles, is_dead, min_thickness, max_thickness
		FROM tools
		WHERE model_type = 'cassette';
	`
)

type CassettesService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewCassettesService(c *shared.Config) *CassettesService {
	return &CassettesService{
		BaseService: shared.NewBaseService(c, "Cassette"),
		mx:          &sync.Mutex{},
	}
}

func (s *CassettesService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateToolTable)
}

func (s *CassettesService) Create(entity shared.ModelTool) (shared.ModelTool, *errors.MasterError) {
	if !entity.IsCassette() {
		return nil, errors.NewValidationError("entity is not a cassette").MasterError()
	}
	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}
	cassette := entity.(*shared.Cassette)

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateCassette,
		sql.Named("position", cassette.Position),
		sql.Named("width", cassette.Width),
		sql.Named("height", cassette.Height),
		sql.Named("type", cassette.Type),
		sql.Named("code", cassette.Code),
		sql.Named("cycles_offset", cassette.CyclesOffset),
		sql.Named("cycles", cassette.Cycles),
		sql.Named("is_dead", cassette.IsDead),
		sql.Named("min_thickness", cassette.MinThickness),
		sql.Named("max_thickness", cassette.MaxThickness),
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	// Store the inserted ID back into the entity
	id, err := r.LastInsertId()
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	if id <= 0 {
		return nil, errors.NewValidationError(
			"invalid ID returned after insert: %v", id,
		).MasterError()
	}
	cassette.ID = shared.EntityID(id)

	return cassette, nil
}

func (s *CassettesService) GetByID(id shared.EntityID) (shared.ModelTool, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetCassetteByID,
		sql.Named("id", id),
	)

	// Scan row into cassette entity
	var c = &shared.Cassette{}
	err := r.Scan(
		&c.ID,
		&c.Position,
		&c.Width,
		&c.Height,
		&c.Type,
		&c.Code,
		&c.CyclesOffset,
		&c.Cycles,
		&c.IsDead,
		&c.MinThickness,
		&c.MaxThickness,
	)
	if err != nil {
		return c, errors.NewMasterError(err, 0)
	}

	return c, nil
}

func (s *CassettesService) Update(entity shared.ModelTool) *errors.MasterError {
	if !entity.IsCassette() {
		return errors.NewValidationError("entity is not a cassette").MasterError()
	}
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}
	cassette := entity.(*shared.Cassette)

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateCassette,
		sql.Named("id", cassette.ID),
		sql.Named("position", cassette.Position),
		sql.Named("width", cassette.Width),
		sql.Named("height", cassette.Height),
		sql.Named("type", cassette.Type),
		sql.Named("code", cassette.Code),
		sql.Named("cycles_offset", cassette.CyclesOffset),
		sql.Named("cycles", cassette.Cycles),
		sql.Named("is_dead", cassette.IsDead),
		sql.Named("min_thickness", cassette.MinThickness),
		sql.Named("max_thickness", cassette.MaxThickness),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CassettesService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteCassette,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CassettesService) List() ([]shared.ModelTool, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListCassettes)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	tools := []shared.ModelTool{}
	for rows.Next() {
		c := &shared.Cassette{}
		err := rows.Scan(
			&c.ID,
			&c.Position,
			&c.Width,
			&c.Height,
			&c.Type,
			&c.Code,
			&c.CyclesOffset,
			&c.Cycles,
			&c.IsDead,
			&c.MinThickness,
			&c.MaxThickness,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		tools = append(tools, c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return tools, nil
}

// Service validation
var _ shared.Service[shared.ModelTool, shared.EntityID] = (*CassettesService)(nil)
