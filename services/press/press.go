package press

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services/shared"
)

const (
	SQLCreatePressTable string = `
		CREATE TABLE IF NOT EXISTS presses (
			id 					INTEGER PRIMARY KEY NOT NULL,
			slot_up 			INTEGER NOT NULL,
			slot_down 			INTEGER NOT NULL,
			last_regeneration 	INTEGER NOT NULL,
			start_cycles 		INTEGER NOT NULL,
			cycles 				INTEGER NOT NULL,
			type 				TEXT NOT NULL
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

type PressService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewPressService(c *shared.Config) *PressService {
	return &PressService{
		BaseService: &shared.BaseService{
			Config: c,
		},

		mx: &sync.Mutex{},
	}
}

func (s *PressService) TableName() string {
	return "presses"
}

func (s *PressService) Setup() *errors.MasterError {
	return s.BaseService.Setup(s.TableName(), SQLCreatePressTable)
}

func (s *PressService) Close() *errors.MasterError {
	// Close the base service connection
	return s.BaseService.Close()
}

func (s *PressService) Create(entity *shared.Press) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
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

	return nil
}

func (s *PressService) Update(entity *shared.Press) *errors.MasterError {
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

func (s *PressService) GetByID(id shared.PressNumber) (*shared.Press, *errors.MasterError) {
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

func (s *PressService) List() ([]*shared.Press, *errors.MasterError) {
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

func (s *PressService) Delete(id shared.PressNumber) *errors.MasterError {
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
var _ shared.Service[*shared.Press, shared.PressNumber] = (*PressService)(nil)
