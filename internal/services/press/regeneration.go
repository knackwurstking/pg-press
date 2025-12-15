package press

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreatePressRegenerationTable string = `
		CREATE TABLE IF NOT EXISTS press_regenerations (
			id 				INTEGER NOT NULL,
			press_number 	INTEGER NOT NULL,
			start 			INTEGER NOT NULL,
			stop 			INTEGER NOT NULL,
			cycles 			INTEGER NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	SQLCreatePressRegeneration string = `
		INSERT INTO press_regenerations (press_number, start, stop, cycles)
		VALUES (:press_number, :start, :stop, :cycles);
	`
	SQLGetPressRegenerationByID string = `
		SELECT id, press_number, start, stop, cycles
		FROM press_regenerations
		WHERE id = :id;
	`
	SQLUpdatePressRegeneration string = `
		UPDATE press_regenerations
		SET press_number = :press_number,
			start = :start,
			stop = :stop,
			cycles = :cycles
		WHERE id = :id;
	`
	SQLDeletePressRegeneration string = `
		DELETE FROM press_regenerations
		WHERE id = :id;
	`
	SQLListPressRegenerations string = `
		SELECT id, press_number, start, stop, cycles
		FROM press_regenerations;
	`
)

// PressRegenerationService provides methods to manage press regenerations
//
// - A press regeneration is a record which will reset a press's cycle count back to zero
// - A regeneration means that the press was broken and got renewed, so the cycle count starts fresh, but this does not matter here

type PressRegenerationService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewPressRegenerationService(c *shared.Config) *PressRegenerationService {
	return &PressRegenerationService{
		BaseService: shared.NewBaseService(c, "PressRegeneration"),
		mx:          &sync.Mutex{},
	}
}

func (s *PressRegenerationService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreatePressRegenerationTable)
}

func (s *PressRegenerationService) Create(entity *shared.PressRegeneration) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreatePressRegeneration,
		sql.Named("press_number", entity.PressNumber),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
		sql.Named("cycles", entity.Cycles),
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

func (s *PressRegenerationService) Update(entity *shared.PressRegeneration) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdatePressRegeneration,
		sql.Named("id", entity.ID),
		sql.Named("press_number", entity.PressNumber),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
		sql.Named("cycles", entity.Cycles),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressRegenerationService) GetByID(id shared.EntityID) (*shared.PressRegeneration, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetPressRegenerationByID,
		sql.Named("id", id),
	)

	// Scan row into press regeneration entity
	var pr = &shared.PressRegeneration{}
	err := r.Scan(
		&pr.ID,
		&pr.PressNumber,
		&pr.Start,
		&pr.Stop,
		&pr.Cycles,
	)
	if err != nil {
		return pr, errors.NewMasterError(err, 0)
	}

	return pr, nil
}

func (s *PressRegenerationService) List() ([]*shared.PressRegeneration, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListPressRegenerations)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	regenerations := []*shared.PressRegeneration{}
	for rows.Next() {
		pr := &shared.PressRegeneration{}
		err := rows.Scan(
			&pr.ID,
			&pr.PressNumber,
			&pr.Start,
			&pr.Stop,
			&pr.Cycles,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		regenerations = append(regenerations, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return regenerations, nil
}

func (s *PressRegenerationService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeletePressRegeneration,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressRegenerationService) Close() *errors.MasterError {
	return s.BaseService.Close()
}

// Service validation
var _ shared.Service[*shared.PressRegeneration, shared.EntityID] = (*PressRegenerationService)(nil)
