// Partial cycles are calculated based on time intervals for each press separately
// The calculation assumes a full cycle takes a predetermined amount of time
package press

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateCycleTable string = `
		CREATE TABLE IF NOT EXISTS press_cycles (
			id           INTEGER NOT NULL,
			tool_id      INTEGER NOT NULL,
			press_number INTEGER NOT NULL,
			cycles       INTEGER NOT NULL, -- Cycles at stop time
			start        INTEGER NOT NULL,
			stop         INTEGER NOT NULL,

			PRIMARY KEY("id")
		);
	`

	SQLCreateCycle string = `
		INSERT INTO press_cycles (tool_id, press_number, cycles, start, stop)
		VALUES (:tool_id, :press_number, :cycles, :start, :stop);
	`

	SQLGetCycleByID string = `
		SELECT id, tool_id, press_number, cycles, start, stop
		FROM press_cycles
		WHERE id = :id;
	`

	SQLUpdateCycle string = `
		UPDATE press_cycles
		SET tool_id      = :tool_id,
			press_number = :press_number,
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
		SELECT id, tool_id, press_number, cycles, start, stop
		FROM press_cycles
		ORDER BY press_number ASC, stop DESC;
	`
)

type PressCyclesService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewPressCyclesService(c *shared.Config) *PressCyclesService {
	return &PressCyclesService{
		BaseService: shared.NewBaseService(c, "Cycle"),
		mx:          &sync.Mutex{},
	}
}

func (s *PressCyclesService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateCycleTable)
}

func (s *PressCyclesService) Create(entity *shared.Cycle) (*shared.Cycle, *errors.MasterError) {
	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateCycle,
		sql.Named("tool_id", entity.ToolID),
		sql.Named("press_number", entity.PressNumber),
		sql.Named("cycles", entity.PressCycles),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
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

func (s *PressCyclesService) Update(entity *shared.Cycle) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateCycle,
		sql.Named("id", entity.ID),
		sql.Named("tool_id", entity.ToolID),
		sql.Named("press_number", entity.PressNumber),
		sql.Named("cycles", entity.PressCycles),
		sql.Named("start", entity.Start),
		sql.Named("stop", entity.Stop),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressCyclesService) GetByID(id shared.EntityID) (*shared.Cycle, *errors.MasterError) {
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
		&c.ToolID,
		&c.PressNumber,
		&c.PressCycles,
		&c.Start,
		&c.Stop,
	)
	if err != nil {
		return c, errors.NewMasterError(err, 0)
	}

	// Calculate partial cycles for this cycle
	c.PartialCycles = CalculatePartialCycles(s.DB(), c)

	return c, nil
}

func (s *PressCyclesService) List() ([]*shared.Cycle, *errors.MasterError) {
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
			&c.ToolID,
			&c.PressNumber,
			&c.PressCycles,
			&c.Start,
			&c.Stop,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		// Calculate partial cycles for each cycle
		c.PartialCycles = CalculatePartialCycles(s.DB(), c)
		cycles = append(cycles, c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return cycles, nil
}

func (s *PressCyclesService) Delete(id shared.EntityID) *errors.MasterError {
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

// -----------------------------------------------------------------------------
// Service Specific Helper Functions
// -----------------------------------------------------------------------------

// CalculatePartialCycles calculates the partial cycles based on the time interval
// This is a placeholder implementation that can be extended with actual cycle times
func CalculatePartialCycles(db *sql.DB, cycle *shared.Cycle) int64 {
	currentCycles := cycle.PressCycles

	// Now we need to get the total press cycles from the last known cycle before the start time of this cycle
	var lastKnownCycles int64 = 0

	row := db.QueryRow(`
		SELECT cycles
		FROM press_cycles
		WHERE press_number = ? AND stop <= ?
		ORDER BY stop DESC
		LIMIT 1;
	`, cycle.PressNumber, cycle.Start)

	err := row.Scan(&lastKnownCycles)
	if err != nil && err != sql.ErrNoRows {
		// In case of error other than no rows, we log and return 0 partial cycles
		return 0
	}

	partial := currentCycles - lastKnownCycles
	return partial
}

// -----------------------------------------------------------------------------
// Interface Validations
// -----------------------------------------------------------------------------

// Service validation
var _ shared.Service[*shared.Cycle, shared.EntityID] = (*PressCyclesService)(nil)
