package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	DBName = "tool"
)

const (
	// TODO: Implement 3 new fields: min_thickness, max_thickness, model_type, update tool and cassette service methods
	SQLCreateToolTable string = `
		CREATE TABLE IF NOT EXISTS tools (
			id					INTEGER NOT NULL,
			position 			INTEGER NOT NULL,
			width 				INTEGER NOT NULL,
			height 				INTEGER NOT NULL,
			type 				TEXT NOT NULL,
			code 				TEXT NOT NULL,
			cycles_offset 		INTEGER NOT NULL DEFAULT 0,
			cycles 				INTEGER NOT NULL DEFAULT 0,
			last_regeneration 	INTEGER NOT NULL DEFAULT 0,
			regenerating 		INTEGER NOT NULL DEFAULT 0,
			is_dead 			INTEGER NOT NULL DEFAULT 0,
			cassette			INTEGER NOT NULL DEFAULT 0,
			min_thickness		REAL NOT NULL,
			max_thickness		REAL NOT NULL,
			model_type			TEXT NOT NULL, -- e.g.: "tool", "cassette"

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	SQLCreateTool string = `
	INSERT INTO tools (position, width, height, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, cassette)
		VALUES (:position, :width, :height, :type, :code, :cycles_offset, :cycles, :last_regeneration, :regenerating, :is_dead, :cassette);
	`
	SQLGetToolByID string = `
		SELECT id, position, width, height, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, cassette
		FROM tools
		WHERE id = :id;
	`
	SQLUpdateTool string = `
		UPDATE tools
		SET position = :position,
			width = :width,
			height = :height,
			type = :type,
			code = :code,
			cycles_offset = :cycles_offset,
			cycles = :cycles,
			last_regeneration = :last_regeneration,
			regenerating = :regenerating,
			is_dead = :is_dead,
			cassette = :cassette
		WHERE id = :id;
	`
	SQLDeleteTool string = `
		UPDATE tools
		SET is_dead = 1
		WHERE id = :id;
	`
	SQLListTools string = `
		SELECT id, position, width, height, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, cassette
		FROM tools;
	`
)

type ToolService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewToolService(c *shared.Config) *ToolService {
	return &ToolService{
		BaseService: shared.NewBaseService(c, "Tool"),
		mx:          &sync.Mutex{},
	}
}

func (s *ToolService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateToolTable)
}

func (s *ToolService) Create(entity *shared.Tool) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateTool,
		sql.Named("position", entity.Position),
		sql.Named("width", entity.Width),
		sql.Named("height", entity.Height),
		sql.Named("type", entity.Type),
		sql.Named("code", entity.Code),
		sql.Named("cycles_offset", entity.CyclesOffset),
		sql.Named("cycles", entity.Cycles),
		sql.Named("last_regeneration", entity.LastRegeneration),
		sql.Named("regenerating", entity.Regenerating),
		sql.Named("is_dead", entity.IsDead),
		sql.Named("cassette", entity.Cassette),
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

func (s *ToolService) GetByID(id shared.EntityID) (*shared.Tool, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetToolByID,
		sql.Named("id", id),
	)

	// Scan row into tool entity
	var t = &shared.Tool{}
	err := r.Scan(
		&t.ID,
		&t.Position,
		&t.Width,
		&t.Height,
		&t.Type,
		&t.Code,
		&t.CyclesOffset,
		&t.Cycles,
		&t.LastRegeneration,
		&t.Regenerating,
		&t.IsDead,
		&t.Cassette,
	)
	if err != nil {
		return t, errors.NewMasterError(err, 0)
	}

	return t, nil
}

func (s *ToolService) Update(entity *shared.Tool) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateTool,
		sql.Named("id", entity.ID),
		sql.Named("position", entity.Position),
		sql.Named("width", entity.Width),
		sql.Named("height", entity.Height),
		sql.Named("type", entity.Type),
		sql.Named("code", entity.Code),
		sql.Named("cycles_offset", entity.CyclesOffset),
		sql.Named("cycles", entity.Cycles),
		sql.Named("last_regeneration", entity.LastRegeneration),
		sql.Named("regenerating", entity.Regenerating),
		sql.Named("is_dead", entity.IsDead),
		sql.Named("cassette", entity.Cassette),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteTool,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolService) List() ([]*shared.Tool, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListTools)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	tools := []*shared.Tool{}
	for rows.Next() {
		t := &shared.Tool{}
		err := rows.Scan(
			&t.ID,
			&t.Position,
			&t.Width,
			&t.Height,
			&t.Type,
			&t.Code,
			&t.CyclesOffset,
			&t.Cycles,
			&t.LastRegeneration,
			&t.Regenerating,
			&t.IsDead,
			&t.Cassette,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		tools = append(tools, t)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return tools, nil
}

// Service validation
var _ shared.Service[*shared.Tool, shared.EntityID] = (*ToolService)(nil)
