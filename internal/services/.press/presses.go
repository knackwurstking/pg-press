package press

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const DBName string = "press"

const (
	SQLCreatePressTable string = `
		CREATE TABLE IF NOT EXISTS presses (
			id 					INTEGER NOT NULL,
			slot_up 			INTEGER NOT NULL,
			slot_down 			INTEGER NOT NULL,
			last_regeneration 	INTEGER NOT NULL,
			start_cycles 		INTEGER NOT NULL,
			cycles 				INTEGER NOT NULL,
			type 				TEXT NOT NULL,

			PRIMARY KEY("id")
		);
	`

	SQLCreatePress string = `
		INSERT INTO presses (id, slot_up, slot_down, last_regeneration, start_cycles, cycles, type)
		VALUES (:id, :slot_up, :slot_down, :last_regeneration, :start_cycles, :cycles, :type);
	`

	SQLGetPressByID string = `
		SELECT id, slot_up, slot_down, last_regeneration, start_cycles, cycles, type
		FROM presses
		WHERE id = :id;
	`

	SQLUpdatePress string = `
		UPDATE presses
		SET slot_up = :slot_up,
			slot_down = :slot_down,
			last_regeneration = :last_regeneration,
			start_cycles = :start_cycles,
			cycles = :cycles,
			type = :type
		WHERE id = :id;
	`

	SQLDeletePress string = `
		DELETE FROM presses
		WHERE id = :id;
	`

	SQLListPresses string = `
		SELECT id, slot_up, slot_down, last_regeneration, start_cycles, cycles, type
		FROM presses;
	`
)

type PressesService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewPressesService(c *shared.Config) *PressesService {
	return &PressesService{
		BaseService: shared.NewBaseService(c, "Press"),
		mx:          &sync.Mutex{},
	}
}

func (s *PressesService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreatePressTable)
}

func (s *PressesService) Close() *errors.MasterError {
	// Close the base service connection
	return s.BaseService.Close()
}

func (s *PressesService) Create(entity *shared.Press) (*shared.Press, *errors.MasterError) {
	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreatePress,
		sql.Named("id", entity.ID),
		sql.Named("slot_up", entity.SlotUp),
		sql.Named("slot_down", entity.SlotDown),
		sql.Named("last_regeneration", entity.LastRegeneration),
		sql.Named("start_cycles", entity.StartCycles),
		sql.Named("cycles", entity.Cycles),
		sql.Named("type", entity.Type),
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

	return entity, nil
}

func (s *PressesService) Update(entity *shared.Press) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdatePress,
		sql.Named("id", entity.ID),
		sql.Named("slot_up", entity.SlotUp),
		sql.Named("slot_down", entity.SlotDown),
		sql.Named("last_regeneration", entity.LastRegeneration),
		sql.Named("start_cycles", entity.StartCycles),
		sql.Named("cycles", entity.Cycles),
		sql.Named("type", entity.Type),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressesService) GetByID(id shared.PressNumber) (*shared.Press, *errors.MasterError) {
	if id < 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetPressByID,
		sql.Named("id", id),
	)

	// Scan row into press entity
	var p = &shared.Press{}
	err := r.Scan(
		&p.ID,
		&p.SlotUp,
		&p.SlotDown,
		&p.LastRegeneration,
		&p.StartCycles,
		&p.Cycles,
		&p.Type,
	)
	if err != nil {
		return p, errors.NewMasterError(err, 0)
	}

	return p, nil
}

func (s *PressesService) List() ([]*shared.Press, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListPresses)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	presses := []*shared.Press{}
	for rows.Next() {
		p := &shared.Press{}
		err := rows.Scan(
			&p.ID,
			&p.SlotUp,
			&p.SlotDown,
			&p.LastRegeneration,
			&p.StartCycles,
			&p.Cycles,
			&p.Type,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		presses = append(presses, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return presses, nil
}

func (s *PressesService) Delete(id shared.PressNumber) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeletePress,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.Press, shared.PressNumber] = (*PressesService)(nil)
