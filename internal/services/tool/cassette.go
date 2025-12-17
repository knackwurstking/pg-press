package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateCassette string = `
		INSERT INTO cassettes (width, height, position, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, min_thickness, max_thickness)
		VALUES (:width, :height, :position, :type, :code, :cycles_offset, :cycles, :last_regeneration, :regenerating, :is_dead, :min_thickness, :max_thickness);
	`
	SQLGetCassetteByID string = `
		SELECT id, width, height, position, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, min_thickness, max_thickness
		FROM cassettes
		WHERE id = :id;
	`
	SQLUpdateCassette string = `
		UPDATE cassettes
		SET width = :width,
			height = :height,
			position = :position,
			type = :type,
			code = :code,
			cycles_offset = :cycles_offset,
			cycles = :cycles,
			last_regeneration = :last_regeneration,
			regenerating = :regenerating,
			is_dead = :is_dead,
			min_thickness = :min_thickness,
			max_thickness = :max_thickness
		WHERE id = :id;
	`
	SQLDeleteCassette string = `
		UPDATE cassettes
		SET is_dead = 1
		WHERE id = :id;
	`
	SQLListCassettes string = `
		SELECT id, width, height, position, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, min_thickness, max_thickness
		FROM cassettes;
	`
)

type CassetteService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewCassetteService(c *shared.Config) *CassetteService {
	return &CassetteService{
		BaseService: shared.NewBaseService(c, "Cassette"),
		mx:          &sync.Mutex{},
	}
}

func (s *CassetteService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateToolTable)
}

func (s *CassetteService) Create(entity *shared.Cassette) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateCassette,
		sql.Named("width", entity.Width),
		sql.Named("height", entity.Height),
		sql.Named("position", entity.Position),
		sql.Named("type", entity.Type),
		sql.Named("code", entity.Code),
		sql.Named("cycles_offset", entity.CyclesOffset),
		sql.Named("cycles", entity.Cycles),
		sql.Named("last_regeneration", entity.LastRegeneration),
		sql.Named("regenerating", entity.Regenerating),
		sql.Named("is_dead", entity.IsDead),
		sql.Named("min_thickness", entity.MinThickness),
		sql.Named("max_thickness", entity.MaxThickness),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	// Store the inserted ID back into the entity
	id, err := r.LastInsertId()
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	if id <= 0 {
		return errors.NewMasterError(
			errors.NewValidationError("invalid ID returned after insert: %v", id), 0)
	}

	entity.ID = shared.EntityID(id)

	return nil
}

func (s *CassetteService) GetByID(id shared.EntityID) (*shared.Cassette, *errors.MasterError) {
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
		&c.Width,
		&c.Height,
		&c.Position,
		&c.Type,
		&c.Code,
		&c.CyclesOffset,
		&c.Cycles,
		&c.LastRegeneration,
		&c.Regenerating,
		&c.IsDead,
		&c.MinThickness,
		&c.MaxThickness,
	)
	if err != nil {
		return c, errors.NewMasterError(err, 0)
	}

	return c, nil
}

func (s *CassetteService) Update(entity *shared.Cassette) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateCassette,
		sql.Named("id", entity.ID),
		sql.Named("width", entity.Width),
		sql.Named("height", entity.Height),
		sql.Named("position", entity.Position),
		sql.Named("type", entity.Type),
		sql.Named("code", entity.Code),
		sql.Named("cycles_offset", entity.CyclesOffset),
		sql.Named("cycles", entity.Cycles),
		sql.Named("last_regeneration", entity.LastRegeneration),
		sql.Named("regenerating", entity.Regenerating),
		sql.Named("is_dead", entity.IsDead),
		sql.Named("min_thickness", entity.MinThickness),
		sql.Named("max_thickness", entity.MaxThickness),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CassetteService) Delete(id shared.EntityID) *errors.MasterError {
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

func (s *CassetteService) List() ([]*shared.Cassette, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListCassettes)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	cassettes := []*shared.Cassette{}
	for rows.Next() {
		c := &shared.Cassette{}
		err := rows.Scan(
			&c.ID,
			&c.Width,
			&c.Height,
			&c.Position,
			&c.Type,
			&c.Code,
			&c.CyclesOffset,
			&c.Cycles,
			&c.LastRegeneration,
			&c.Regenerating,
			&c.IsDead,
			&c.MinThickness,
			&c.MaxThickness,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		cassettes = append(cassettes, c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return cassettes, nil
}

// Service validation
var _ shared.Service[*shared.Cassette, shared.EntityID] = (*CassetteService)(nil)
