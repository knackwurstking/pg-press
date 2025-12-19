package tool

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

var (
	SQLCreateLowerTool = fmt.Sprintf(`
		INSERT INTO tools (position, width, height, type, code, cycles_offset, cycles, is_dead, cassette, model_type)
		VALUES (%d, :width, :height, :type, :code, :cycles_offset, :cycles, :is_dead, :cassette, 'tool');
	`, shared.SlotLower)

	SQLGetLowerToolByID = fmt.Sprintf(`
		SELECT id, position, width, height, type, code, cycles_offset, cycles, is_dead, cassette
		FROM tools
		WHERE id = :id AND model_type = 'tool' AND position = %d;
	`, shared.SlotLower)

	SQLUpdateLowerTool = fmt.Sprintf(`
		UPDATE tools
		SET position = :position,
			width = :width,
			height = :height,
			type = :type,
			code = :code,
			cycles_offset = :cycles_offset,
			cycles = :cycles,
			is_dead = :is_dead,
			cassette = :cassette
		WHERE id = :id AND model_type = 'tool' AND position = %d;
	`, shared.SlotLower)

	SQLDeleteLowerTool = fmt.Sprintf(`
		DELETE FROM tools
		WHERE id = :id AND model_type = 'tool' AND position = %d;
	`, shared.SlotLower)

	SQLListLowerTools = fmt.Sprintf(`
		SELECT id, position, width, height, type, code, cycles_offset, cycles, is_dead, cassette
		FROM tools
		WHERE model_type = 'tool' AND position = %d;
	`, shared.SlotLower)
)

type LowerToolsService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewLowerToolsService(c *shared.Config) *LowerToolsService {
	return &LowerToolsService{
		BaseService: shared.NewBaseService(c, "Tool"),
		mx:          &sync.Mutex{},
	}
}

func (s *LowerToolsService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateToolTable)
}

func (s *LowerToolsService) Create(entity shared.ModelTool) (shared.ModelTool, *errors.MasterError) {
	if entity.IsCassette() {
		return nil, errors.NewValidationError(
			"cannot create cassette with LowerToolsService",
		).MasterError()
	}
	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}
	tool := entity.(*shared.Tool)
	if tool.Position != shared.SlotLower {
		return nil, errors.NewValidationError("invalid position for lower tool: %d", tool.Position).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateLowerTool,
		sql.Named("width", tool.Width),
		sql.Named("height", tool.Height),
		sql.Named("type", tool.Type),
		sql.Named("code", tool.Code),
		sql.Named("cycles_offset", tool.CyclesOffset),
		sql.Named("cycles", tool.Cycles),
		sql.Named("is_dead", tool.IsDead),
		sql.Named("cassette", tool.Cassette),
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

	tool.ID = shared.EntityID(id)

	return tool, nil
}

func (s *LowerToolsService) GetByID(id shared.EntityID) (shared.ModelTool, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetLowerToolByID,
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
		&t.IsDead,
		&t.Cassette,
	)
	if err != nil {
		return t, errors.NewMasterError(err, 0)
	}

	return t, nil
}

func (s *LowerToolsService) Update(entity shared.ModelTool) *errors.MasterError {
	if entity.IsCassette() {
		return errors.NewValidationError(
			"cannot update cassette with LowerToolsService",
		).MasterError()
	}
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}
	tool := entity.(*shared.Tool)
	if tool.Position != shared.SlotLower {
		return errors.NewValidationError(
			"invalid position for lower tool: %d", tool.Position,
		).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateLowerTool,
		sql.Named("id", tool.ID),
		sql.Named("width", tool.Width),
		sql.Named("height", tool.Height),
		sql.Named("type", tool.Type),
		sql.Named("code", tool.Code),
		sql.Named("cycles_offset", tool.CyclesOffset),
		sql.Named("cycles", tool.Cycles),
		sql.Named("is_dead", tool.IsDead),
		sql.Named("cassette", tool.Cassette),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *LowerToolsService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteLowerTool,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *LowerToolsService) List() ([]shared.ModelTool, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListLowerTools)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	tools := []shared.ModelTool{}
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
var _ shared.Service[shared.ModelTool, shared.EntityID] = (*LowerToolsService)(nil)
