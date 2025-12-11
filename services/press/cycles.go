package press

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services/shared"
)

const (
	SQLCreateCycleTable string = `
		CREATE TABLE IF NOT EXISTS press_cycles (
			id           INTEGER NOT NULL,
			press_number INTEGER NOT NULL,
			cycles       INTEGER NOT NULL,
			start        INTEGER NOT NULL,
			stop         INTEGER NOT NULL,

			PRIMARY KEY("id")
		);
	`
	SQLCreateCycle string = `
		INSERT INTO press_cycles (press_number, cycles, start, stop)
		VALUES (:press_number, :cycles, :start, :stop);
	`
	SQLGetCycleByID string = `
		SELECT id, press_number, cycles, start, stop
		FROM press_cycles
		WHERE id = :id;
	`
	SQLUpdateCycle string = `
		UPDATE press_cycles
		SET press_number = :press_number,
			cycles       = :cycles,
			start        = :start,
			stop         = :stop
		WHERE id = :id;
	`
	SQLDeleteCycle string = `
		DELETE FROM press_cycles
		WHERE id = :id;
	`
	SQLListCycles string = `
		SELECT id, press_number, cycles, start, stop
		FROM press_cycles;
	`
)

type CycleService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewCycleService(c *shared.Config) *CycleService {
	return &CycleService{
		BaseService: &shared.BaseService{
			Config: c,
		},

		mx: &sync.Mutex{},
	}
}

func (s *CycleService) TableName() string {
	return "press_cycles"
}

func (s *CycleService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, s.TableName(), SQLCreateCycleTable)
}

func (s *CycleService) Create(entity *shared.Cycle) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateCycle,
		sql.Named("press_number", entity.PressNumber),
		sql.Named("cycles", entity.Cycles),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
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

func (s *CycleService) Update(entity *shared.Cycle) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateCycle,
		sql.Named("id", entity.ID),
		sql.Named("press_number", entity.PressNumber),
		sql.Named("cycles", entity.Cycles),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CycleService) GetByID(id shared.EntityID) (*shared.Cycle, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetCycleByID,
		sql.Named("id", id),
	)

	// Scan row into cycle entity
	var c = &shared.Cycle{}
	err := r.Scan(
		&c.ID,
		&c.PressNumber,
		&c.Cycles,
		&c.Start,
		&c.Stop,
	)
	if err != nil {
		return c, errors.NewMasterError(err, 0)
	}

	return c, nil
}

func (s *CycleService) List() ([]*shared.Cycle, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListCycles)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	cycles := []*shared.Cycle{}
	for rows.Next() {
		c := &shared.Cycle{}
		err := rows.Scan(
			&c.ID,
			&c.PressNumber,
			&c.Cycles,
			&c.Start,
			&c.Stop,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		cycles = append(cycles, c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return cycles, nil
}

func (s *CycleService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteCycle,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.Cycle, shared.EntityID] = (*CycleService)(nil)
