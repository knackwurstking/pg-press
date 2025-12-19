package tool

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateToolRegenerationTable string = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id 		INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			start 	INTEGER NOT NULL,
			stop 	INTEGER NOT NULL,
			cycles 	INTEGER NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	SQLCreateToolRegeneration string = `
		INSERT INTO tool_regenerations (tool_id, start, stop, cycles)
		VALUES (:tool_id, :start, :stop, :cycles);
	`

	SQLGetToolRegenerationByID string = `
		SELECT id, tool_id, start, stop, cycles
		FROM tool_regenerations
		WHERE id = :id;
	`

	SQLUpdateToolRegeneration string = `
		UPDATE tool_regenerations
		SET tool_id = :tool_id,
			start = :start,
			stop = :stop,
			cycles = :cycles
		WHERE id = :id;
	`

	SQLDeleteToolRegeneration string = `
		DELETE FROM tool_regenerations
		WHERE id = :id;
	`

	SQLListToolRegenerations string = `
		SELECT id, tool_id, start, stop, cycles
		FROM tool_regenerations
		ORDER BY start DESC;
	`
)

type ToolRegenerationsService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewToolRegenerationsService(c *shared.Config) *ToolRegenerationsService {
	return &ToolRegenerationsService{
		BaseService: shared.NewBaseService(c, "ToolRegeneration"),
		mx:          &sync.Mutex{},
	}
}

func (s *ToolRegenerationsService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateToolRegenerationTable)
}

func (s *ToolRegenerationsService) Create(entity *shared.ToolRegeneration) (*shared.ToolRegeneration, *errors.MasterError) {
	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// Check if there's already an ongoing regeneration for this tool (where stop = 0)
	verr = s.checkOngoingRegeneration(entity)
	if verr != nil {
		return nil, verr.MasterError()
	}

	r, err := s.DB().Exec(SQLCreateToolRegeneration,
		sql.Named("tool_id", entity.ToolID),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
		sql.Named("cycles", entity.Cycles),
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

	entity.ID = shared.EntityID(id)

	return entity, nil
}

func (s *ToolRegenerationsService) Update(entity *shared.ToolRegeneration) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateToolRegeneration,
		sql.Named("tool_id", entity.ToolID),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
		sql.Named("cycles", entity.Cycles),
		sql.Named("id", entity.ID),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolRegenerationsService) GetByID(id shared.EntityID) (*shared.ToolRegeneration, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetToolRegenerationByID,
		sql.Named("id", id),
	)

	// Scan row into tool regeneration entity
	var tr = &shared.ToolRegeneration{}
	err := r.Scan(
		&tr.ID,
		&tr.ToolID,
		&tr.Start,
		&tr.Stop,
		&tr.Cycles,
	)
	if err != nil {
		return tr, errors.NewMasterError(err, 0)
	}

	return tr, nil
}

func (s *ToolRegenerationsService) List() ([]*shared.ToolRegeneration, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListToolRegenerations)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	regenerations := []*shared.ToolRegeneration{}
	for rows.Next() {
		tr := &shared.ToolRegeneration{}
		err := rows.Scan(
			&tr.ID,
			&tr.ToolID,
			&tr.Start,
			&tr.Stop,
			&tr.Cycles,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		regenerations = append(regenerations, tr)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return regenerations, nil
}

func (s *ToolRegenerationsService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteToolRegeneration,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolRegenerationsService) checkOngoingRegeneration(entity *shared.ToolRegeneration) *errors.ValidationError {
	row := s.DB().QueryRow(
		`
			SELECT id FROM tool_regenerations
			WHERE tool_id = :tool_id AND (start > 0 AND (stop = 0 OR stop > ?))
			ORDER BY stop DESC
			LIMIT 1;
		`,
		entity.Start,
		sql.Named("tool_id", entity.ToolID),
	)
	if row != nil && row.Err() != nil && row.Err() != sql.ErrNoRows {
		return errors.NewValidationError(
			"there is already an ongoing regeneration for tool ID %v: %v",
			entity.ToolID, row.Err(),
		)
	}
	return nil
}

// Service validation
var _ shared.Service[*shared.ToolRegeneration, shared.EntityID] = (*ToolRegenerationsService)(nil)
